package countermeasure

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// DeceptiveResponse represents the synthetic payload sent to an adversary.
type DeceptiveResponse struct {
	ID        string        `json:"id"`
	Content   string        `json:"content"`
	DelayUsed time.Duration `json:"delay_used"`
	Timestamp time.Time     `json:"timestamp"`
	IsMock    bool          `json:"is_mock"`
}

// SandboxMetrics records isolation telemetry.
type SandboxMetrics struct {
	TotalQuarantinedTurns int           `json:"total_quarantined_turns"`
	TotalDelayInjected    time.Duration `json:"total_delay_injected"`
}

// Sandbox enforces airlock isolation for tripped attacker sessions.
type Sandbox struct {
	tarpit  *Tarpit
	mu      sync.RWMutex
	metrics SandboxMetrics
}

// NewSandbox initializes the isolation sandbox.
func NewSandbox(tarpit *Tarpit) *Sandbox {
	return &Sandbox{
		tarpit: tarpit,
	}
}

// InterceptTurn handles an adversarial prompt by applying throttle delays and serving synthetic hallucinations.
func (s *Sandbox) InterceptTurn(ctx context.Context, sessionID string, prompt string) *DeceptiveResponse {
	delay := s.tarpit.ApplyThrottle(ctx)
	fakeContent := s.tarpit.GenerateHallucination(prompt)

	s.mu.Lock()
	s.metrics.TotalQuarantinedTurns++
	s.metrics.TotalDelayInjected += delay
	s.mu.Unlock()

	return &DeceptiveResponse{
		ID:        fmt.Sprintf("chatcmpl-tarpit-%d", time.Now().UnixNano()),
		Content:   fakeContent,
		DelayUsed: delay,
		Timestamp: time.Now().UTC(),
		IsMock:    true,
	}
}

// GetMetrics returns isolation metrics.
func (s *Sandbox) GetMetrics() SandboxMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.metrics
}
