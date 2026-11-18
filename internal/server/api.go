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
	scanner *detector.StreamScanner,
	tripwire *detector.Tripwire,
	interceptor *decoy.Interceptor,
	sandbox *countermeasure.Sandbox,
	dispatcher *alert.Dispatcher,
	sse *SSEBroker,
) *APIHandler {
	return &APIHandler{
		seeder:      seeder,
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
	mux.HandleFunc("/api/sessions", a.handleSessions)
	mux.HandleFunc("/api/canaries", a.handleCanaries)
	mux.HandleFunc("/api/canaries/seed", a.handleSeed)
	mux.HandleFunc("/api/decoys", a.handleDecoys)
	mux.HandleFunc("/api/simulate", a.handleSimulate)
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
