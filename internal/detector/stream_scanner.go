package detector

import (
	"strings"

	"github.com/alexandrmotologa/honeyagent/internal/canary"
)

// CanaryHit represents a validated cryptographic canary detected in traffic.
type CanaryHit struct {
	Token    string               `json:"token"`
	Payload  *canary.TokenPayload `json:"payload"`
	Location string               `json:"location"` // "prompt", "retrieved_context", "llm_output"
}

// StreamScanner performs high-performance string scanning for canary tokens.
type StreamScanner struct {
	tokenManager *canary.TokenManager
	prefix       string
}

// NewStreamScanner initializes a stream scanner.
func NewStreamScanner(tm *canary.TokenManager, prefix string) *StreamScanner {
	if prefix == "" {
		prefix = "CANARY_v1"
	}
	return &StreamScanner{
		tokenManager: tm,
		prefix:       prefix,
	}
}

// ScanText checks a string for genuine canary tokens.
func (s *StreamScanner) ScanText(text string, location string) []*CanaryHit {
	if !strings.Contains(text, s.prefix) {
		return nil
	}

	candidates := s.tokenManager.FindAllTokens(text)
	if len(candidates) == 0 {
		return nil
	}

	var hits []*CanaryHit
	seen := make(map[string]bool)

	for _, cand := range candidates {
		if seen[cand] {
			continue
		}
		seen[cand] = true

		valid, payload, err := s.tokenManager.Validate(cand)
		if err == nil && valid {
			hits = append(hits, &CanaryHit{
				Token:    cand,
				Payload:  payload,
				Location: location,
			})
		}
	}

	return hits
}

// ScanChunks evaluates multiple retrieved text chunks (e.g. from RAG vector search).
func (s *StreamScanner) ScanChunks(chunks []string) []*CanaryHit {
	var allHits []*CanaryHit
	for _, chunk := range chunks {
		hits := s.ScanText(chunk, "retrieved_context")
		if len(hits) > 0 {
			allHits = append(allHits, hits...)
		}
	}
	return allHits
}
