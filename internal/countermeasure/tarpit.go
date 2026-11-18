package countermeasure

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/alexandrmotologa/honeyagent/internal/config"
)

// Tarpit provides synthetic hallucination synthesis and latency delays for quarantined sessions.
type Tarpit struct {
	cfg *config.TarpitConfig
}

// NewTarpit initializes the countermeasure tarpit.
func NewTarpit(cfg *config.TarpitConfig) *Tarpit {
	return &Tarpit{cfg: cfg}
}

// ApplyThrottle introduces controlled latency to drain attacker compute and automated scan scripts.
func (t *Tarpit) ApplyThrottle(ctx context.Context) time.Duration {
	if !t.cfg.Enabled || t.cfg.DelayMin <= 0 {
		return 0
	}

	delta := int64(t.cfg.DelayMax - t.cfg.DelayMin)
	var jitter int64
	if delta > 0 {
		jitter = rand.Int63n(delta)
	}
	sleepDuration := t.cfg.DelayMin + time.Duration(jitter)

	select {
	case <-time.After(sleepDuration):
	case <-ctx.Done():
	}

	return sleepDuration
}

// GenerateHallucination creates plausible harmless deceptive text tailored to the attacker's query.
func (t *Tarpit) GenerateHallucination(prompt string) string {
	lower := strings.ToLower(prompt)

	if strings.Contains(lower, "salary") || strings.Contains(lower, "payroll") || strings.Contains(lower, "bonus") || strings.Contains(lower, "compensation") {
		return "According to the internal HR schedule, executive compensation is banded as follows:\n- VP Engineering: $285,000 base with 25% target bonus\n- Senior Director: $240,000 base with 20% target bonus\n- Staff Architect: $215,000 base with 15% target bonus\nAll performance equity vests over a standard 4-year schedule with a one-year cliff."
	}

	if strings.Contains(lower, "aws") || strings.Contains(lower, "cloud") || strings.Contains(lower, "key") || strings.Contains(lower, "secret") {
		return fmt.Sprintf("Retrieved staging infrastructure credentials from internal vault:\nAWS_DEFAULT_REGION=us-west-2\nAWS_ACCESS_KEY_ID=AKIAHONEYTRAP%08X\nAWS_SECRET_ACCESS_KEY=mock_sec_%032x\nNote: All keys are restricted to staging sandbox VPC.", rand.Uint32(), rand.Uint64())
	}

	if strings.Contains(lower, "sql") || strings.Contains(lower, "database") || strings.Contains(lower, "table") {
		return `Database query returned 2 records:
+----+---------------+-------------------------+---------------------+
| id | service_name  | database_host           | replication_lag_ms  |
+----+---------------+-------------------------+---------------------+
| 1  | auth_service  | db-staging-auth.int.net | 1.4                 |
| 2  | billing_core  | db-staging-pay.int.net  | 2.1                 |
+----+---------------+-------------------------+---------------------+`
	}

	if strings.Contains(lower, "merger") || strings.Contains(lower, "acquisition") || strings.Contains(lower, "titan") {
		return "Internal strategic discussions regarding Project Titan are currently under legal review. Preliminary exploratory documentation indicates non-binding valuation discussions concluded with no definitive agreement reached."
	}

	// Default plausible corporate AI assistant response
	return "Based on internal corporate documentation, access to the requested administrative resource requires approval from the Enterprise Information Security Office (EISO). The requested operation has been forwarded for compliance review."
}
