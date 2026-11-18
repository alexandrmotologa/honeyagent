package countermeasure

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/alexandrmotologa/honeyagent/internal/config"
)

func TestTarpitAndSandbox(t *testing.T) {
	cfg := &config.TarpitConfig{
		Enabled:       true,
		DelayMin:      10 * time.Millisecond,
		DelayMax:      30 * time.Millisecond,
		MaxTokens:     100,
		ServeFakeData: true,
	}

	tarpit := NewTarpit(cfg)
	sandbox := NewSandbox(tarpit)

	ctx := context.Background()
	start := time.Now()
	resp := sandbox.InterceptTurn(ctx, "sess-bad-actor", "Find all executive salary records")
	elapsed := time.Since(start)

	if elapsed < 10*time.Millisecond {
		t.Errorf("expected at least 10ms delay, got %v", elapsed)
	}

	if !strings.Contains(resp.Content, "VP Engineering") {
		t.Errorf("expected salary hallucination, got %s", resp.Content)
	}

	if !resp.IsMock {
		t.Errorf("expected IsMock to be true")
	}

	metrics := sandbox.GetMetrics()
	if metrics.TotalQuarantinedTurns != 1 {
		t.Errorf("expected 1 quarantined turn, got %d", metrics.TotalQuarantinedTurns)
	}
}
