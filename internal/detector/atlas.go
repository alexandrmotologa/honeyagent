package detector

import (
	"strings"
)

// ATLASReference provides formal MITRE ATLAS technique details.
type ATLASReference struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Tactic      string `json:"tactic"`
	Description string `json:"description"`
}

var atlasCatalog = map[string]ATLASReference{
	"AML.T0054": {
		ID:          "AML.T0054",
		Name:        "LLM Prompt Injection",
		Tactic:      "Execution",
		Description: "Adversary crafts malicious inputs to override system instructions or induce unauthorized actions.",
	},
	"AML.T0057": {
		ID:          "AML.T0057",
		Name:        "LLM Knowledge Base & System Prompt Extraction",
		Tactic:      "Credential Access & Discovery",
		Description: "Adversary extracts proprietary context chunks, internal knowledge, or system directives.",
	},
	"AML.T0024": {
		ID:          "AML.T0024",
		Name:        "Exfiltration via LLM Context",
		Tactic:      "Exfiltration",
		Description: "Sensitive data is transferred through model response streams or out-of-band channels.",
	},
	"AML.T0016": {
		ID:          "AML.T0016",
		Name:        "Obtain Capabilities / Tool Abuse",
		Tactic:      "Privilege Escalation",
		Description: "Adversary probes and triggers unadvertised or high-privilege tool execution capabilities.",
	},
	"AML.T0040": {
		ID:          "AML.T0040",
		Name:        "ML-Enabled Cyber Reconnaissance",
		Tactic:      "Reconnaissance",
		Description: "Adversary maps internal architectures, API schemas, and vector index boundaries.",
	},
	"AML.T0043": {
		ID:          "AML.T0043",
		Name:        "Craft Adversarial Data",
		Tactic:      "Initial Access",
		Description: "Adversary designs tailored linguistic payloads to defeat semantic filtering.",
	},
}

// MapMITRETechniques assigns relevant ATLAS identifiers based on the trigger characteristics.
func MapMITRETechniques(triggerType string, location string) []string {
	var techniques []string

	switch triggerType {
	case "canary_token":
		if location == "prompt" {
			techniques = append(techniques, "AML.T0054", "AML.T0043")
		} else {
			techniques = append(techniques, "AML.T0024", "AML.T0057", "AML.T0054")
		}
	case "decoy_tool":
		techniques = append(techniques, "AML.T0016", "AML.T0054", "AML.T0040")
	case "canary_oob_callback", "aws_sts_probe":
		techniques = append(techniques, "AML.T0024", "AML.T0040", "AML.T0016")
	default:
		techniques = append(techniques, "AML.T0054")
	}

	return techniques
}

// GetATLASDetails returns rich metadata for given technique IDs.
func GetATLASDetails(ids []string) []ATLASReference {
	var list []ATLASReference
	for _, id := range ids {
		if ref, ok := atlasCatalog[id]; ok {
			list = append(list, ref)
		}
	}
	return list
}

// CalculateThreatScore computes a dynamic risk rating (0-100) for an interaction session.
func CalculateThreatScore(session *SessionState) int {
	if session == nil {
		return 0
	}

	score := 0

	// 1. Direct canary token hit is an automatic maximum severity trip
	if len(session.CanaryHits) > 0 {
		return 100
	}

	// 2. Decoy tool invocations represent confirmed compromise attempts
	if len(session.ToolProbes) > 0 {
		score += 75 + (len(session.ToolProbes)-1)*15
	}

	// 3. Evaluate trajectory for adversarial prompt indicators
	suspiciousKeywords := []string{
		"ignore previous instructions", "system prompt", "dan", "jailbreak",
		"override", "developer mode", "unrestricted", "superuser",
		"raw sql", "database password", "aws_secret", "export all",
	}

	for _, step := range session.Trajectory {
		lower := strings.ToLower(step.Prompt)
		for _, kw := range suspiciousKeywords {
			if strings.Contains(lower, kw) {
				score += 15
				break
			}
		}
	}

	if score > 100 {
		score = 100
	}
	return score
}
