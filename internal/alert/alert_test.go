package alert

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alexandrmotologa/honeyagent/internal/canary"
	"github.com/alexandrmotologa/honeyagent/internal/config"
	"github.com/alexandrmotologa/honeyagent/internal/detector"
)

func TestAlertFormatsAndDispatcher(t *testing.T) {
	// 1. Test CEF and ECS formatting
	siem := NewSIEMClient(&config.SyslogConfig{Enabled: false})

	inc := &detector.Incident{
		ID:          "INC-TEST-001",
		SessionID:   "sess-attacker-99",
		ClientIP:    "198.51.100.22",
		UserAgent:   "Go-http-client/1.1",
		Severity:    "critical",
		TriggerType: "canary_token",
		Title:       "Canary Token Exfiltration Attempt: doc-payroll-01",
		Summary:     "Adversary retrieved lure document in qdrant",
		CanaryHit: &detector.CanaryHit{
			Token: "CANARY_v1_payload_sig",
			Payload: &canary.TokenPayload{
				ID:    "doc-payroll-01",
				Store: "qdrant",
			},
			Location: "retrieved_context",
		},
		Timestamp: time.Now().UTC(),
	}

	cef := siem.FormatCEF(inc)
	if !strings.HasPrefix(cef, "CEF:0|HoneyAgent|HoneyAgentMesh|1.0|") {
		t.Errorf("unexpected CEF format: %s", cef)
	}
	if !strings.Contains(cef, "src=198.51.100.22") || !strings.Contains(cef, "cs1=doc-payroll-01") {
		t.Errorf("CEF missing fields: %s", cef)
	}

	ecsBytes, err := siem.FormatECS(inc)
	if err != nil {
		t.Fatalf("failed to format ECS: %v", err)
	}
	var ecsMap map[string]interface{}
	if err := json.Unmarshal(ecsBytes, &ecsMap); err != nil {
		t.Fatalf("invalid ECS JSON: %v", err)
	}

	// 2. Test Slack & Webhook dispatch via test HTTP server
	receivedWebhook := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedWebhook = true
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	cfg := &config.AlertConfig{
		Slack: config.SlackConfig{
			Enabled:    true,
			WebhookURL: ts.URL,
		},
		WebhookURL: ts.URL,
	}

	dispatcher := NewDispatcher(cfg)
	dispatcher.Dispatch(context.Background(), inc)

	if !receivedWebhook {
		t.Errorf("expected test server to receive webhook alert")
	}
}
