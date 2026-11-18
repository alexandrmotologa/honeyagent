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

// SlackClient dispatches real-time incident alert cards to Slack channels.
type SlackClient struct {
	cfg        *config.SlackConfig
	httpClient *http.Client
}

// NewSlackClient creates a Slack alert client.
func NewSlackClient(cfg *config.SlackConfig) *SlackClient {
	return &SlackClient{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// Send formats and transmits a rich Slack card.
func (s *SlackClient) Send(ctx context.Context, inc *detector.Incident) error {
	if !s.cfg.Enabled || s.cfg.WebhookURL == "" {
		return nil
	}

	canaryOrTool := "Unknown"
	if inc.CanaryHit != nil && inc.CanaryHit.Payload != nil {
		canaryOrTool = fmt.Sprintf("Canary Document: `%s` (Store: %s)", inc.CanaryHit.Payload.ID, inc.CanaryHit.Payload.Store)
	} else if inc.ToolProbe != nil {
		canaryOrTool = fmt.Sprintf("Decoy Tool: `%s`", inc.ToolProbe.ToolName)
	}

	payload := map[string]interface{}{
		"text": fmt.Sprintf("🚨 *[HONEYAGENT TRIPWIRE]* %s", inc.Title),
		"blocks": []map[string]interface{}{
			{
				"type": "header",
				"text": map[string]string{
					"type":  "plain_text",
					"text":  "🚨 Critical Security Alert: HoneyAgent Tripwire Triggered",
					"emoji": "true",
				},
			},
			{
				"type": "section",
				"fields": []map[string]string{
					{"type": "mrkdwn", "text": fmt.Sprintf("*Incident ID:*\n`%s`", inc.ID)},
					{"type": "mrkdwn", "text": fmt.Sprintf("*Severity:*\n*%s*", inc.Severity)},
					{"type": "mrkdwn", "text": fmt.Sprintf("*Attacker IP:*\n`%s`", inc.ClientIP)},
					{"type": "mrkdwn", "text": fmt.Sprintf("*Session ID:*\n`%s`", inc.SessionID)},
				},
			},
			{
				"type": "section",
				"text": map[string]string{
					"type": "mrkdwn",
					"text": fmt.Sprintf("*Trigger Target:*\n%s\n\n*Forensic Summary:*\n%s", canaryOrTool, inc.Summary),
				},
			},
			{
				"type": "context",
				"elements": []map[string]string{
					{"type": "mrkdwn", "text": fmt.Sprintf("Action: Session routed to Tarpit Sandbox | Time: %s", inc.Timestamp.Format(time.RFC3339))},
				},
			},
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.cfg.WebhookURL, bytes.NewReader(data))
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
		return fmt.Errorf("slack webhook returned HTTP %d", resp.StatusCode)
	}

	return nil
}
