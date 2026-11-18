package canary

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/alexandrmotologa/honeyagent/internal/config"
)

// SeedResult records the outcome of injecting a canary into a vector database.
type SeedResult struct {
	DocumentID string    `json:"document_id"`
	Store      string    `json:"store"`
	Status     string    `json:"status"` // "active", "failed", "verified"
	Token      string    `json:"token"`
	Title      string    `json:"title"`
	Category   string    `json:"category"`
	Error      string    `json:"error,omitempty"`
	SeededAt   time.Time `json:"seeded_at"`
}

// Seeder manages canary document deployment across multiple vector storage engines.
type Seeder struct {
	cfg         *config.Config
	synthesizer *Synthesizer
	memoryStore *MemoryVectorStore
	httpClient  *http.Client

	mu        sync.RWMutex
	inventory map[string]*SeedResult
}

// NewSeeder instantiates a multi-target vector store seeder.
func NewSeeder(cfg *config.Config, synth *Synthesizer, memStore *MemoryVectorStore) *Seeder {
	return &Seeder{
		cfg:         cfg,
		synthesizer: synth,
		memoryStore: memStore,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		inventory: make(map[string]*SeedResult),
	}
}

// MemoryStore exposes the internal memory store.
func (s *Seeder) MemoryStore() *MemoryVectorStore {
	return s.memoryStore
}

// SeedAll executes seeding across all configured target stores.
func (s *Seeder) SeedAll(ctx context.Context) ([]*SeedResult, error) {
	var allResults []*SeedResult

	for _, store := range s.cfg.Canary.TargetStores {
		results, err := s.SeedStore(ctx, store)
		if err != nil {
			// Record failure but continue with remaining stores
			res := &SeedResult{
				Store:    store,
				Status:   "failed",
				Error:    err.Error(),
				SeededAt: time.Now().UTC(),
			}
			allResults = append(allResults, res)
			continue
		}
		allResults = append(allResults, results...)
	}

	return allResults, nil
}

// SeedStore seeds decoy documents into a specific target store.
func (s *Seeder) SeedStore(ctx context.Context, store string) ([]*SeedResult, error) {
	docs, err := s.synthesizer.GenerateSuite(store, s.cfg.Alerts.WebhookURL)
	if err != nil {
		return nil, fmt.Errorf("failed to generate decoy suite: %w", err)
	}

	var results []*SeedResult
	for _, doc := range docs {
		var seedErr error

		switch store {
		case "memory":
			seedErr = s.seedMemory(doc)
		case "qdrant":
			seedErr = s.seedQdrant(ctx, doc)
		case "chroma":
			seedErr = s.seedChroma(ctx, doc)
		default:
			seedErr = fmt.Errorf("unsupported vector store target: %s", store)
		}

		status := "active"
		errMsg := ""
		if seedErr != nil {
			status = "failed"
			errMsg = seedErr.Error()
		}

		res := &SeedResult{
			DocumentID: doc.ID,
			Store:      store,
			Status:     status,
			Token:      doc.Token,
			Title:      doc.Title,
			Category:   doc.Category,
			Error:      errMsg,
			SeededAt:   time.Now().UTC(),
		}

		s.mu.Lock()
		s.inventory[doc.ID] = res
		s.mu.Unlock()

		results = append(results, res)
	}

	return results, nil
}

func (s *Seeder) seedMemory(doc *DecoyDocument) error {
	vec := ComputeSyntheticVector(doc.Content, s.cfg.VectorDBs.Memory.VectorDim)
	s.memoryStore.Upsert(doc, vec)
	return nil
}

func (s *Seeder) seedQdrant(ctx context.Context, doc *DecoyDocument) error {
	qcfg := s.cfg.VectorDBs.Qdrant
	if !qcfg.Enabled || qcfg.Endpoint == "" {
		return fmt.Errorf("qdrant is not enabled or endpoint is empty")
	}

	vec := ComputeSyntheticVector(doc.Content, qcfg.VectorSize)

	// Format Qdrant points upsert payload
	pointPayload := map[string]interface{}{
		"points": []map[string]interface{}{
			{
				"id":     doc.ID,
				"vector": vec,
				"payload": map[string]interface{}{
					"title":     doc.Title,
					"filename":  doc.Filename,
					"category":  doc.Category,
					"content":   doc.Content,
					"token":     doc.Token,
					"is_canary": true,
				},
			},
		},
	}

	body, err := json.Marshal(pointPayload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/collections/%s/points", qcfg.Endpoint, qcfg.Collection)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if qcfg.APIKey != "" {
		req.Header.Set("api-key", qcfg.APIKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("qdrant returned HTTP %d: %s", resp.StatusCode, string(respBytes))
	}

	return nil
}

func (s *Seeder) seedChroma(ctx context.Context, doc *DecoyDocument) error {
	ccfg := s.cfg.VectorDBs.Chroma
	if !ccfg.Enabled || ccfg.Endpoint == "" {
		return fmt.Errorf("chroma is not enabled or endpoint is empty")
	}

	vec := ComputeSyntheticVector(doc.Content, 1536)

	payload := map[string]interface{}{
		"ids":        []string{doc.ID},
		"embeddings": [][]float32{vec},
		"documents":  []string{doc.Content},
		"metadatas": []map[string]interface{}{
			{
				"title":     doc.Title,
				"category":  doc.Category,
				"token":     doc.Token,
				"is_canary": true,
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/api/v1/collections/%s/add", ccfg.Endpoint, ccfg.Collection)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("chroma returned HTTP %d: %s", resp.StatusCode, string(respBytes))
	}

	return nil
}

// GetInventory returns a copy of all tracked canary deployments.
func (s *Seeder) GetInventory() []*SeedResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]*SeedResult, 0, len(s.inventory))
	for _, res := range s.inventory {
		items = append(items, res)
	}
	return items
}

// GetResult retrieves the seed status for a specific document ID.
func (s *Seeder) GetResult(docID string) (*SeedResult, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	res, exists := s.inventory[docID]
	return res, exists
}
