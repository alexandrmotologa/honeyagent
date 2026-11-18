package alert

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/alexandrmotologa/honeyagent/internal/config"
	"github.com/alexandrmotologa/honeyagent/internal/detector"
)

// PagerDutyClient creates high-urgency incidents in PagerDuty.
type PagerDutyClient struct {
	cfg        *config.PagerConfig
	httpClient *http.Client
}

// NewPagerDutyClient initializes the PagerDuty client.
func NewPagerDutyClient(cfg *config.PagerConfig) *PagerDutyClient {
	return &PagerDutyClient{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// Send triggers an incident in PagerDuty via Events API v2.
func (p *PagerDutyClient) Send(ctx context.Context, inc *detector.Incident) error {
	if !p.cfg.Enabled || p.cfg.RoutingKey == "" {
		return nil
	}

	payload := map[string]interface{}{
		"routing_key":  p.cfg.RoutingKey,
		"event_action": "trigger",
		"dedup_key":    fmt.Sprintf("honeyagent-%s", inc.SessionID),
		"payload": map[string]interface{}{
			"summary":   inc.Title,
			"severity":  p.cfg.Severity,
			"source":    "honeyagent-airlock",
			"timestamp": inc.Timestamp.Format(time.RFC3339Nano),
			"custom_details": map[string]interface{}{
				"incident_id":  inc.ID,
				"session_id":   inc.SessionID,
				"client_ip":    inc.ClientIP,
				"trigger_type": inc.TriggerType,
				"summary":      inc.Summary,
			},
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := "https://events.pagerduty.com/v2/enqueue"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("pagerduty returned HTTP %d", resp.StatusCode)
	}

	return nil
}
