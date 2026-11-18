package alert

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/alexandrmotologa/honeyagent/internal/config"
	"github.com/alexandrmotologa/honeyagent/internal/detector"
)

// SIEMClient formats and transmits security events to enterprise SIEM platforms.
type SIEMClient struct {
	cfg *config.SyslogConfig
}

// NewSIEMClient initializes the SIEM exporter.
func NewSIEMClient(cfg *config.SyslogConfig) *SIEMClient {
	return &SIEMClient{cfg: cfg}
}

// FormatCEF converts an incident into an ArcSight Common Event Format string.
func (s *SIEMClient) FormatCEF(inc *detector.Incident) string {
	severity := 10
	if inc.Severity != "critical" {
		severity = 7
	}

	canaryID := "none"
	if inc.CanaryHit != nil && inc.CanaryHit.Payload != nil {
		canaryID = inc.CanaryHit.Payload.ID
	}
	toolName := "none"
	if inc.ToolProbe != nil {
		toolName = inc.ToolProbe.ToolName
	}

	// CEF:Version|Device Vendor|Device Product|Device Version|Device Event Class ID|Name|Severity|[Extension]
	return fmt.Sprintf("CEF:0|HoneyAgent|HoneyAgentMesh|1.0|%s|%s|%d|src=%s suser=%s cs1=%s cs1Label=CanaryID cs2=%s cs2Label=DecoyTool msg=%s",
		inc.TriggerType,
		escapeCEF(inc.Title),
		severity,
		inc.ClientIP,
		inc.SessionID,
		canaryID,
		toolName,
		escapeCEF(inc.Summary),
	)
}

// FormatECS converts an incident into an Elastic Common Schema (ECS) JSON structure.
func (s *SIEMClient) FormatECS(inc *detector.Incident) ([]byte, error) {
	record := map[string]interface{}{
		"@timestamp": inc.Timestamp.Format(time.RFC3339Nano),
		"event": map[string]interface{}{
			"kind":     "alert",
			"category": []string{"intrusion_detection"},
			"type":     []string{"indicator"},
			"outcome":  "failure",
			"severity": 10,
			"action":   "session_quarantine",
			"id":       inc.ID,
		},
		"source": map[string]interface{}{
			"ip": inc.ClientIP,
		},
		"user": map[string]interface{}{
			"id": inc.SessionID,
		},
		"user_agent": map[string]interface{}{
			"original": inc.UserAgent,
		},
		"message": inc.Summary,
		"rule": map[string]interface{}{
			"name":        inc.Title,
			"description": "HoneyAgent autonomous deception tripwire",
		},
	}

	if inc.CanaryHit != nil && inc.CanaryHit.Payload != nil {
		record["threat"] = map[string]interface{}{
			"indicator": map[string]interface{}{
				"type": "canary_token",
				"name": inc.CanaryHit.Payload.ID,
			},
		}
	} else if inc.ToolProbe != nil {
		record["threat"] = map[string]interface{}{
			"indicator": map[string]interface{}{
				"type": "decoy_tool",
				"name": inc.ToolProbe.ToolName,
			},
		}
	}

	return json.Marshal(record)
}

// Send dispatches the formatted event over Syslog network socket if enabled.
func (s *SIEMClient) Send(inc *detector.Incident) error {
	if !s.cfg.Enabled || s.cfg.Address == "" {
		return nil
	}

	cefMsg := s.FormatCEF(inc)
	payload := fmt.Sprintf("<134>%s %s %s: %s\n",
		time.Now().Format(time.RFC3339),
		"honeyagent",
		s.cfg.Tag,
		cefMsg,
	)

	conn, err := net.DialTimeout(s.cfg.Network, s.cfg.Address, 3*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to syslog %s: %w", s.cfg.Address, err)
	}
	defer conn.Close()

	_, err = conn.Write([]byte(payload))
	return err
}

func escapeCEF(val string) string {
	val = strings.ReplaceAll(val, "\\", "\\\\")
	val = strings.ReplaceAll(val, "|", "\\|")
	val = strings.ReplaceAll(val, "\n", "\\n")
	return val
}
