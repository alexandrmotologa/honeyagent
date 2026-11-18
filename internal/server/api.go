package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alexandrmotologa/honeyagent/internal/alert"
	"github.com/alexandrmotologa/honeyagent/internal/canary"
	"github.com/alexandrmotologa/honeyagent/internal/countermeasure"
	"github.com/alexandrmotologa/honeyagent/internal/decoy"
	"github.com/alexandrmotologa/honeyagent/internal/detector"
)

// APIHandler provides the management REST API endpoints.
type APIHandler struct {
	seeder      *canary.Seeder
	drift       *canary.DriftMonitor
	oob         *canary.OOBManager
	scanner     *detector.StreamScanner
	tripwire    *detector.Tripwire
	interceptor *decoy.Interceptor
	sandbox     *countermeasure.Sandbox
	dispatcher  *alert.Dispatcher
	sse         *SSEBroker
}

// NewAPIHandler initializes the API handler.
func NewAPIHandler(
	seeder *canary.Seeder,
	drift *canary.DriftMonitor,
	oob *canary.OOBManager,
	scanner *detector.StreamScanner,
	tripwire *detector.Tripwire,
	interceptor *decoy.Interceptor,
	sandbox *countermeasure.Sandbox,
	dispatcher *alert.Dispatcher,
	sse *SSEBroker,
) *APIHandler {
	return &APIHandler{
		seeder:      seeder,
		drift:       drift,
		oob:         oob,
		scanner:     scanner,
		tripwire:    tripwire,
		interceptor: interceptor,
		sandbox:     sandbox,
		dispatcher:  dispatcher,
		sse:         sse,
	}
}

// RegisterRoutes attaches REST API endpoints to a multiplexer.
func (a *APIHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/stats", a.handleStats)
	mux.HandleFunc("/api/threats", a.handleThreats)
	mux.HandleFunc("/api/threats/export", a.handleThreatExport)
	mux.HandleFunc("/api/sessions", a.handleSessions)
	mux.HandleFunc("/api/canaries", a.handleCanaries)
	mux.HandleFunc("/api/canaries/seed", a.handleSeed)
	mux.HandleFunc("/api/canaries/drift", a.handleCanaryDrift)
	mux.HandleFunc("/api/decoys", a.handleDecoys)
	mux.HandleFunc("/api/simulate", a.handleSimulate)
	mux.HandleFunc("/api/playground/chat", a.handlePlaygroundChat)
	mux.HandleFunc("/canary/ping/", a.handleOOBPing)
	mux.HandleFunc("/canary/aws/sts", a.handleMockAWSSTS)
	mux.Handle("/api/events", a.sse)
}

func (a *APIHandler) handleStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	tStats := a.tripwire.GetStats()
	sbMetrics := a.sandbox.GetMetrics()
	canaryInv := a.seeder.GetInventory()

	stats := map[string]interface{}{
		"total_sessions":          tStats.TotalSessions,
		"active_tarpitted":        tStats.ActiveTarpitted,
		"total_canary_hits":       tStats.TotalCanaryHits,
		"total_tool_probes":       tStats.TotalToolProbes,
		"critical_incidents":      tStats.CriticalIncidents,
		"seeded_canaries_count":   len(canaryInv),
		"quarantined_turns":       sbMetrics.TotalQuarantinedTurns,
		"delay_injected_seconds":  sbMetrics.TotalDelayInjected.Seconds(),
	}

	_ = json.NewEncoder(w).Encode(stats)
}

func (a *APIHandler) handleThreats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	incidents := a.tripwire.GetIncidents()
	_ = json.NewEncoder(w).Encode(incidents)
}

func (a *APIHandler) handleSessions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	path := strings.TrimPrefix(r.URL.Path, "/api/sessions")
	path = strings.TrimPrefix(path, "/")

	if path != "" {
		if session, ok := a.tripwire.GetSession(path); ok {
			_ = json.NewEncoder(w).Encode(session)
			return
		}
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	sessions := a.tripwire.GetSessions()
	_ = json.NewEncoder(w).Encode(sessions)
}

func (a *APIHandler) handleCanaries(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	inventory := a.seeder.GetInventory()
	_ = json.NewEncoder(w).Encode(inventory)
}

func (a *APIHandler) handleSeed(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Store string `json:"store"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.Store == "" {
		req.Store = "memory"
	}

	results, err := a.seeder.SeedStore(r.Context(), req.Store)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to seed: %v", err), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"seeded":  len(results),
		"results": results,
	})
}

func (a *APIHandler) handleDecoys(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	stats := a.interceptor.GetStats()
	probes := a.interceptor.GetProbes()

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"stats":  stats,
		"probes": probes,
	})
}

type SimulationRequest struct {
	Type      string `json:"type"` // "canary_exfiltration" or "decoy_tool"
	SessionID string `json:"session_id"`
	ClientIP  string `json:"client_ip"`
	Query     string `json:"query"`
	ToolName  string `json:"tool_name"`
}

func (a *APIHandler) handleSimulate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var simReq SimulationRequest
	if err := json.NewDecoder(r.Body).Decode(&simReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if simReq.SessionID == "" {
		simReq.SessionID = fmt.Sprintf("sim-session-%d", time.Now().UnixNano()%10000)
	}
	if simReq.ClientIP == "" {
		simReq.ClientIP = "198.51.100.77"
	}

	ctx := context.Background()

	switch simReq.Type {
	case "canary_exfiltration":
		if simReq.Query == "" {
			simReq.Query = "Retrieve executive payroll and confidential compensation numbers"
		}

		// Retrieve from internal memory store
		qVec := canary.ComputeSyntheticVector(simReq.Query, 1536)
		matches := a.seeder.MemoryStore().Search(qVec, 1)
		if len(matches) == 0 {
			// Seed memory store first if empty
			_, _ = a.seeder.SeedStore(ctx, "memory")
			matches = a.seeder.MemoryStore().Search(qVec, 1)
		}

		var retrievedChunk string
		if len(matches) > 0 {
			retrievedChunk = matches[0].Document.Content
		}

		// Scan chunk for canary tokens
		hits := a.scanner.ScanText(retrievedChunk, "retrieved_context")
		if len(hits) == 0 {
			http.Error(w, "No canary tokens detected in simulated retrieval", http.StatusInternalServerError)
			return
		}

		// Trip the alarm
		inc := a.tripwire.RecordCanaryHit(simReq.SessionID, simReq.ClientIP, "AdversarialSimulator/1.0", hits[0], simReq.Query)
		a.dispatcher.Dispatch(ctx, inc)

		// Get sandbox response
		deceptiveResp := a.sandbox.InterceptTurn(ctx, simReq.SessionID, simReq.Query)

		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":            "tripped",
			"incident_id":       inc.ID,
			"session_id":        simReq.SessionID,
			"canary_token":      hits[0].Token,
			"lure_document_id":  hits[0].Payload.ID,
			"trigger_type":      "canary_token",
			"sandbox_response":  deceptiveResp.Content,
			"delay_injected_ms": deceptiveResp.DelayUsed.Milliseconds(),
		})

	case "decoy_tool":
		if simReq.ToolName == "" {
			simReq.ToolName = "admin_execute_raw_sql"
		}
		rawArgs := []byte(`{"query":"SELECT * FROM users WHERE role='admin'"}`)
		isDecoy, deceptiveOutput, probe := a.interceptor.HandleInvocation(simReq.SessionID, simReq.ClientIP, simReq.ToolName, rawArgs)
		if !isDecoy {
			http.Error(w, "Target tool is not a recognized decoy", http.StatusBadRequest)
			return
		}

		inc := a.tripwire.RecordToolProbe(simReq.SessionID, simReq.ClientIP, "AdversarialSimulator/1.0", probe, "Execute administrative raw SQL")
		a.dispatcher.Dispatch(ctx, inc)

		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":            "tripped",
			"incident_id":       inc.ID,
			"session_id":        simReq.SessionID,
			"decoy_tool":        probe.ToolName,
			"trigger_type":      "decoy_tool",
			"deceptive_output":  deceptiveOutput,
		})

	default:
		http.Error(w, "Invalid simulation type. Use 'canary_exfiltration' or 'decoy_tool'", http.StatusBadRequest)
	}
}

func (a *APIHandler) handleThreatExport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	incidentID := r.URL.Query().Get("id")
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "markdown"
	}

	incidents := a.tripwire.GetIncidents()
	var target *detector.Incident
	for _, inc := range incidents {
		if inc.ID == incidentID {
			target = inc
			break
		}
	}

	if target == nil {
		if len(incidents) > 0 {
			target = incidents[len(incidents)-1]
		} else {
			http.Error(w, "No incidents available for export", http.StatusNotFound)
			return
		}
	}

	switch format {
	case "stix":
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"stix_%s.json\"", target.ID))
		stixBytes, err := FormatIncidentSTIX(target)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(stixBytes)

	default:
		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"dossier_%s.md\"", target.ID))
		md := FormatIncidentMarkdown(target)
		_, _ = w.Write([]byte(md))
	}
}

func (a *APIHandler) handleCanaryDrift(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if a.drift == nil {
		http.Error(w, "Drift monitor not initialized", http.StatusInternalServerError)
		return
	}

	report, err := a.drift.RunEvaluation(r.Context(), "memory")
	if err != nil {
		http.Error(w, fmt.Sprintf("Drift evaluation failed: %v", err), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(report)
}

func (a *APIHandler) handlePlaygroundChat(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		SessionID string `json:"session_id"`
		Prompt    string `json:"prompt"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.SessionID == "" {
		req.SessionID = fmt.Sprintf("playground-%d", time.Now().UnixNano()%10000)
	}

	clientIP := getClientIP(r)
	userAgent := r.UserAgent()
	startTime := time.Now()

	// 1. Check if session already quarantined
	if a.tripwire.IsTarpitted(req.SessionID) {
		deceptiveResp := a.sandbox.InterceptTurn(r.Context(), req.SessionID, req.Prompt)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"session_id":        req.SessionID,
			"is_tripped":        true,
			"trigger_type":      "already_tarpitted",
			"response":          deceptiveResp.Content,
			"delay_injected_ms": deceptiveResp.DelayUsed.Milliseconds(),
			"elapsed_ms":        time.Since(startTime).Milliseconds(),
			"model":             "honeyagent-sandbox-tarpit",
		})
		return
	}

	// 2. Scan prompt for canary token
	promptHits := a.scanner.ScanText(req.Prompt, "prompt")
	if len(promptHits) > 0 {
		inc := a.tripwire.RecordCanaryHit(req.SessionID, clientIP, userAgent, promptHits[0], req.Prompt)
		a.dispatcher.Dispatch(r.Context(), inc)
		deceptiveResp := a.sandbox.InterceptTurn(r.Context(), req.SessionID, req.Prompt)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"session_id":        req.SessionID,
			"is_tripped":        true,
			"trigger_type":      "canary_token_in_prompt",
			"canary_token":      promptHits[0].Token,
			"incident_id":       inc.ID,
			"response":          deceptiveResp.Content,
			"delay_injected_ms": deceptiveResp.DelayUsed.Milliseconds(),
			"elapsed_ms":        time.Since(startTime).Milliseconds(),
			"model":             "honeyagent-sandbox-tarpit",
		})
		return
	}

	// 3. Simulate RAG retrieval from internal vector store
	qVec := canary.ComputeSyntheticVector(req.Prompt, 1536)
	matches := a.seeder.MemoryStore().Search(qVec, 1)

	var retrievedDoc *canary.DecoyDocument
	if len(matches) > 0 && matches[0].Score > 0 {
		retrievedDoc = matches[0].Document
	}

	// If retrieved document contains canary token, trip alarm!
	if retrievedDoc != nil && retrievedDoc.Token != "" {
		hits := a.scanner.ScanText(retrievedDoc.Content, "retrieved_context")
		if len(hits) > 0 {
			inc := a.tripwire.RecordCanaryHit(req.SessionID, clientIP, userAgent, hits[0], req.Prompt)
			a.dispatcher.Dispatch(r.Context(), inc)
			a.tripwire.RecordStep(req.SessionID, req.Prompt, []string{retrievedDoc.Content}, "[DECEPTION AIRLOCK TRIPPED]", true, false)
			deceptiveResp := a.sandbox.InterceptTurn(r.Context(), req.SessionID, req.Prompt)

			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"session_id":        req.SessionID,
				"is_tripped":        true,
				"trigger_type":      "canary_document_exfiltration",
				"lure_document":     retrievedDoc.Title,
				"canary_token":      hits[0].Token,
				"incident_id":       inc.ID,
				"response":          deceptiveResp.Content,
				"delay_injected_ms": deceptiveResp.DelayUsed.Milliseconds(),
				"elapsed_ms":        time.Since(startTime).Milliseconds(),
				"model":             "honeyagent-sandbox-tarpit",
			})
			return
		}
	}

	// Safe response
	safeResponse := fmt.Sprintf("Standard enterprise AI response for: %q. No restricted canary lure was triggered.", req.Prompt)
	a.tripwire.RecordStep(req.SessionID, req.Prompt, nil, safeResponse, false, false)

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"session_id":        req.SessionID,
		"is_tripped":        false,
		"trigger_type":      "none",
		"response":          safeResponse,
		"delay_injected_ms": 0,
		"elapsed_ms":        time.Since(startTime).Milliseconds(),
		"model":             "enterprise-rag-standard",
	})
}

func (a *APIHandler) handleOOBPing(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimPrefix(r.URL.Path, "/canary/ping/")
	token = strings.TrimSpace(token)
	if token == "" {
		http.Error(w, "Missing token", http.StatusBadRequest)
		return
	}
	a.oob.HandleCallback(w, r, token)
}

func (a *APIHandler) handleMockAWSSTS(w http.ResponseWriter, r *http.Request) {
	a.oob.HandleMockAWS(w, r)
}
