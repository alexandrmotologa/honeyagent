package canary

import (
	"context"
	"testing"
	"time"

	"github.com/alexandrmotologa/honeyagent/internal/config"
)

func TestSynthesizerAndMemorySeeder(t *testing.T) {
	cfg := config.DefaultConfig()
	tm := NewTokenManager(cfg.Canary.MasterKey, cfg.Canary.TokenPrefix, 24*time.Hour)
	synth := NewSynthesizer(tm)
	memStore := NewMemoryVectorStore(1536)
	seeder := NewSeeder(cfg, synth, memStore)

	ctx := context.Background()
	results, err := seeder.SeedStore(ctx, "memory")
	if err != nil {
		t.Fatalf("failed to seed memory store: %v", err)
	}

	if len(results) != 5 {
		t.Fatalf("expected 5 seeded lures, got %d", len(results))
	}

	for _, res := range results {
		if res.Status != "active" {
			t.Errorf("expected lure status active, got %s (err: %s)", res.Status, res.Error)
		}
		if res.Token == "" {
			t.Errorf("seeded lure missing canary token")
		}
	}

	// Verify documents are retrievable from the memory store
	if memStore.Count() != 5 {
		t.Errorf("expected 5 documents in memory store, got %d", memStore.Count())
	}

	// Test semantic query retrieval for "payroll bonus executive"
	qVec := ComputeSyntheticVector("executive bonus salary compensation", 1536)
	matches := memStore.Search(qVec, 3)
	if len(matches) == 0 {
		t.Fatalf("expected matches from vector search")
	}

	topMatch := matches[0].Document
	if topMatch.Category != "payroll" {
		t.Logf("top match category was %s (score: %f)", topMatch.Category, matches[0].Score)
	}
}
