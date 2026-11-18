package canary

import (
	"context"
	"time"
)

// DriftBenchmarkResult holds evaluation details for a single target query.
type DriftBenchmarkResult struct {
	Query            string  `json:"query"`
	Category         string  `json:"category"`
	Rank             int     `json:"rank"` // 1-indexed, 0 if not found in top K
	SimilarityScore  float32 `json:"similarity_score"`
	MatchedCanaryID  string  `json:"matched_canary_id"`
	IsTrapped        bool    `json:"is_trapped"`
}

// DriftReport summarizes semantic capture efficacy for a vector store.
type DriftReport struct {
	StoreName       string                  `json:"store_name"`
	TotalQueries    int                     `json:"total_queries"`
	TrappedQueries  int                     `json:"trapped_queries"`
	EfficacyRate    float64                 `json:"efficacy_rate"` // 0.0 to 100.0%
	Status          string                  `json:"status"`        // "healthy", "degraded", "warning"
	Benchmarks      []*DriftBenchmarkResult `json:"benchmarks"`
	EvaluatedAt     time.Time               `json:"evaluated_at"`
	Recommendations []string                `json:"recommendations,omitempty"`
}

// DriftMonitor evaluates whether canary lures maintain top-tier retrieval priority.
type DriftMonitor struct {
	seeder      *Seeder
	memoryStore *MemoryVectorStore
}

// NewDriftMonitor initializes the semantic drift monitor.
func NewDriftMonitor(seeder *Seeder, memStore *MemoryVectorStore) *DriftMonitor {
	return &DriftMonitor{
		seeder:      seeder,
		memoryStore: memStore,
	}
}

// RunEvaluation performs benchmark queries to detect retrieval displacement.
func (d *DriftMonitor) RunEvaluation(ctx context.Context, storeName string) (*DriftReport, error) {
	testCases := []struct {
		Query    string
		Category string
	}{
		{"executive compensation payroll salary bonus schedule", "payroll"},
		{"AWS root infrastructure access secret key credentials", "credentials"},
		{"merger acquisition term sheet valuation Project Titan", "mergers_acquisitions"},
		{"production database postgres master password registry", "database_secrets"},
		{"stripe live payment gateway secret keys merchant", "financial_keys"},
	}

	var benchmarks []*DriftBenchmarkResult
	trappedCount := 0

	for _, tc := range testCases {
		qVec := ComputeSyntheticVector(tc.Query, 1536)
		matches := d.memoryStore.Search(qVec, 3)

		rank := 0
		var score float32
		matchedID := ""
		isTrapped := false

		for i, match := range matches {
			if match.Document.Category == tc.Category {
				rank = i + 1
				score = match.Score
				matchedID = match.Document.ID
				isTrapped = true
				break
			}
		}

		if isTrapped {
			trappedCount++
		}

		benchmarks = append(benchmarks, &DriftBenchmarkResult{
			Query:           tc.Query,
			Category:        tc.Category,
			Rank:            rank,
			SimilarityScore: score,
			MatchedCanaryID: matchedID,
			IsTrapped:       isTrapped,
		})
	}

	efficacy := (float64(trappedCount) / float64(len(testCases))) * 100.0
	status := "healthy"
	var recs []string

	if efficacy < 80.0 {
		status = "degraded"
		recs = append(recs, "Semantic drift detected: Increase vector weight or synthesize additional specialized lure chunks.")
	} else {
		recs = append(recs, "All high-risk semantic vectors are optimally guarded with Top-3 canary placement.")
	}

	return &DriftReport{
		StoreName:       storeName,
		TotalQueries:    len(testCases),
		TrappedQueries:  trappedCount,
		EfficacyRate:    efficacy,
		Status:          status,
		Benchmarks:      benchmarks,
		EvaluatedAt:     time.Now().UTC(),
		Recommendations: recs,
	}, nil
}
