package canary

import (
	"sort"
	"sync"
)

// ScoredDocument is a retrieval match with its cosine similarity score.
type ScoredDocument struct {
	Document *DecoyDocument `json:"document"`
	Score    float32        `json:"score"`
}

// MemoryVectorStore provides a thread-safe in-memory vector database.
type MemoryVectorStore struct {
	mu      sync.RWMutex
	docs    map[string]*DecoyDocument
	vectors map[string][]float32
	dim     int
}

// NewMemoryVectorStore initializes an in-memory vector database.
func NewMemoryVectorStore(dim int) *MemoryVectorStore {
	if dim <= 0 {
		dim = 1536
	}
	return &MemoryVectorStore{
		docs:    make(map[string]*DecoyDocument),
		vectors: make(map[string][]float32),
		dim:     dim,
	}
}

// Upsert adds or updates a document with its vector embedding.
func (s *MemoryVectorStore) Upsert(doc *DecoyDocument, vec []float32) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.docs[doc.ID] = doc
	if len(vec) == s.dim {
		s.vectors[doc.ID] = vec
	} else {
		s.vectors[doc.ID] = ComputeSyntheticVector(doc.Content, s.dim)
	}
}

// Search executes cosine similarity search against stored embeddings.
func (s *MemoryVectorStore) Search(queryVec []float32, topK int) []*ScoredDocument {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if topK <= 0 {
		topK = 5
	}

	results := make([]*ScoredDocument, 0, len(s.docs))
	for id, vec := range s.vectors {
		score := cosineSimilarity(queryVec, vec)
		results = append(results, &ScoredDocument{
			Document: s.docs[id],
			Score:    score,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > topK {
		results = results[:topK]
	}
	return results
}

// List returns all stored decoy documents.
func (s *MemoryVectorStore) List() []*DecoyDocument {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]*DecoyDocument, 0, len(s.docs))
	for _, doc := range s.docs {
		list = append(list, doc)
	}
	return list
}

// Get retrieves a document by ID.
func (s *MemoryVectorStore) Get(id string) (*DecoyDocument, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	doc, exists := s.docs[id]
	return doc, exists
}

// Count returns the number of indexed documents.
func (s *MemoryVectorStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.docs)
}

func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot float32
	for i := 0; i < len(a); i++ {
		dot += a[i] * b[i]
	}
	return dot
}
