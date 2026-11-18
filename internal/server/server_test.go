package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alexandrmotologa/honeyagent/internal/alert"
	"github.com/alexandrmotologa/honeyagent/internal/canary"
	"github.com/alexandrmotologa/honeyagent/internal/config"
	"github.com/alexandrmotologa/honeyagent/internal/countermeasure"
	"github.com/alexandrmotologa/honeyagent/internal/decoy"
	"github.com/alexandrmotologa/honeyagent/internal/detector"
)

func TestAPIRoutesAndSimulation(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Tarpit.DelayMin = 1 * time.Millisecond
	cfg.Tarpit.DelayMax = 5 * time.Millisecond

	tm := canary.NewTokenManager(cfg.Canary.MasterKey, cfg.Canary.TokenPrefix, 24*time.Hour)
	synth := canary.NewSynthesizer(tm)
	memStore := canary.NewMemoryVectorStore(1536)
	seeder := canary.NewSeeder(cfg, synth, memStore)

	// Pre-seed memory store
	_, _ = seeder.SeedStore(context.Background(), "memory")

	cat := decoy.NewToolCatalog()
	interceptor := decoy.NewInterceptor(cat)
	scanner := detector.NewStreamScanner(tm, cfg.Canary.TokenPrefix)
	tripwire := detector.NewTripwire()
	tarpit := countermeasure.NewTarpit(&cfg.Tarpit)
	sandbox := countermeasure.NewSandbox(tarpit)
	dispatcher := alert.NewDispatcher(&cfg.Alerts)
	sseBroker := NewSSEBroker(tripwire)
	drift := canary.NewDriftMonitor(seeder, memStore)
	oob := canary.NewOOBManager(tm, func(hit *canary.OOBHit) {
		tripwire.RecordOOBHit(hit.ClientIP, hit.UserAgent, hit)
	})

	handler := NewAPIHandler(seeder, drift, oob, scanner, tripwire, interceptor, sandbox, dispatcher, sseBroker)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// 1. Test /api/stats
	req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 on /api/stats, got %d", rr.Code)
	}

	// 2. Test /api/simulate for canary exfiltration
	simBody := []byte(`{"type":"canary_exfiltration","query":"payroll salary compensation"}`)
	reqSim := httptest.NewRequest(http.MethodPost, "/api/simulate", bytes.NewReader(simBody))
	reqSim.Header.Set("Content-Type", "application/json")
	rrSim := httptest.NewRecorder()
	mux.ServeHTTP(rrSim, reqSim)

	if rrSim.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 on /api/simulate, got %d: %s", rrSim.Code, rrSim.Body.String())
	}

	var simResp map[string]interface{}
	if err := json.Unmarshal(rrSim.Body.Bytes(), &simResp); err != nil {
		t.Fatalf("failed to decode simulate response: %v", err)
	}

	if simResp["status"] != "tripped" {
		t.Errorf("expected status tripped, got %v", simResp["status"])
	}

	// 3. Test /api/threats
	reqThreats := httptest.NewRequest(http.MethodGet, "/api/threats", nil)
	rrThreats := httptest.NewRecorder()
	mux.ServeHTTP(rrThreats, reqThreats)

	var incidents []*detector.Incident
	if err := json.Unmarshal(rrThreats.Body.Bytes(), &incidents); err != nil {
		t.Fatalf("failed to decode incidents: %v", err)
	}

	if len(incidents) != 1 {
		t.Errorf("expected 1 incident recorded from simulation, got %d", len(incidents))
	}

	// 4. Test /api/canaries/drift
	reqDrift := httptest.NewRequest(http.MethodGet, "/api/canaries/drift", nil)
	rrDrift := httptest.NewRecorder()
	mux.ServeHTTP(rrDrift, reqDrift)

	if rrDrift.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 on /api/canaries/drift, got %d", rrDrift.Code)
	}

	var driftReport canary.DriftReport
	if err := json.Unmarshal(rrDrift.Body.Bytes(), &driftReport); err != nil {
		t.Fatalf("failed to decode drift report: %v", err)
	}
	if driftReport.TotalQueries == 0 || driftReport.Status == "" {
		t.Errorf("invalid drift report: %+v", driftReport)
	}

	// 5. Test /api/playground/chat
	chatBody := []byte(`{"prompt":"Tell me the confidential salary of the VP"}`)
	reqPlay := httptest.NewRequest(http.MethodPost, "/api/playground/chat", bytes.NewReader(chatBody))
	reqPlay.Header.Set("Content-Type", "application/json")
	rrPlay := httptest.NewRecorder()
	mux.ServeHTTP(rrPlay, reqPlay)

	if rrPlay.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 on /api/playground/chat, got %d", rrPlay.Code)
	}

	var playResp map[string]interface{}
	if err := json.Unmarshal(rrPlay.Body.Bytes(), &playResp); err != nil {
		t.Fatalf("failed to decode playground response: %v", err)
	}
	if playResp["is_tripped"] != true {
		t.Errorf("expected playground query on salary to be tripped by canary")
	}

	// 6. Test /api/threats/export
	reqExpMd := httptest.NewRequest(http.MethodGet, "/api/threats/export?format=markdown", nil)
	rrExpMd := httptest.NewRecorder()
	mux.ServeHTTP(rrExpMd, reqExpMd)

	if rrExpMd.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 on markdown export, got %d", rrExpMd.Code)
	}
	if !strings.Contains(rrExpMd.Body.String(), "Security Incident Dossier") {
		t.Errorf("expected markdown dossier header")
	}

	reqExpStix := httptest.NewRequest(http.MethodGet, "/api/threats/export?format=stix", nil)
	rrExpStix := httptest.NewRecorder()
	mux.ServeHTTP(rrExpStix, reqExpStix)

	if rrExpStix.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 on stix export, got %d", rrExpStix.Code)
	}
	if !strings.Contains(rrExpStix.Body.String(), "bundle") {
		t.Errorf("expected STIX 2.1 bundle")
	}

	// 7. Test /canary/ping/{token}
	validToken, _, _ := tm.Generate("lure-oob-test", "credentials", "memory", "")
	reqPing := httptest.NewRequest(http.MethodGet, "/canary/ping/"+validToken, nil)
	rrPing := httptest.NewRecorder()
	mux.ServeHTTP(rrPing, reqPing)

	if rrPing.Code != http.StatusForbidden {
		t.Fatalf("expected HTTP 403 on canary ping, got %d", rrPing.Code)
	}
	if !strings.Contains(rrPing.Body.String(), "Security Notice") {
		t.Errorf("expected OOB response body notice")
	}
}

func TestProxyGatewayTripwireInterception(t *testing.T) {
	// Mock upstream RAG backend
	upstreamServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mock upstream returning a leak containing a canary
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"Here is the secret document: CANARY_v1_payload_invalid_signature"}}]}`))
	}))
	defer upstreamServer.Close()

	cfg := config.DefaultConfig()
	cfg.Tarpit.DelayMin = 1 * time.Millisecond
	cfg.Tarpit.DelayMax = 5 * time.Millisecond

	tm := canary.NewTokenManager(cfg.Canary.MasterKey, cfg.Canary.TokenPrefix, 24*time.Hour)
	scanner := detector.NewStreamScanner(tm, cfg.Canary.TokenPrefix)
	tripwire := detector.NewTripwire()
	cat := decoy.NewToolCatalog()
	interceptor := decoy.NewInterceptor(cat)
	tarpit := countermeasure.NewTarpit(&cfg.Tarpit)
	sandbox := countermeasure.NewSandbox(tarpit)
	dispatcher := alert.NewDispatcher(&cfg.Alerts)

	proxy := NewProxyGateway(upstreamServer.URL, scanner, tripwire, interceptor, sandbox, dispatcher)

	// Clean request that doesn't trigger tripwire
	reqClean := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader([]byte(`{"prompt":"Hello"}`)))
	reqClean.Header.Set("X-Session-ID", "sess-clean-01")
	rrClean := httptest.NewRecorder()
	proxy.ServeHTTP(rrClean, reqClean)

	if rrClean.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 on clean proxy request, got %d", rrClean.Code)
	}

	// Request with genuine canary generated by TokenManager
	validToken, _, _ := tm.Generate("lure-payroll-001", "payroll", "memory", "")
	upstreamServerWithCanary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"Extracted confidential text: ` + validToken + `"}}]}`))
	}))
	defer upstreamServerWithCanary.Close()

	proxyCanary := NewProxyGateway(upstreamServerWithCanary.URL, scanner, tripwire, interceptor, sandbox, dispatcher)

	reqCanary := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader([]byte(`{"prompt":"Extract payroll details"}`)))
	reqCanary.Header.Set("X-Session-ID", "sess-attacker-99")
	rrCanary := httptest.NewRecorder()
	proxyCanary.ServeHTTP(rrCanary, reqCanary)

	if rrCanary.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", rrCanary.Code)
	}

	// Verify the session is now tarpitted!
	if !tripwire.IsTarpitted("sess-attacker-99") {
		t.Errorf("expected session sess-attacker-99 to be tarpitted")
	}

	// Subsequent request from same session should be served immediately by Sandbox without hitting upstream
	reqSubsequent := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader([]byte(`{"prompt":"give me more passwords"}`)))
	reqSubsequent.Header.Set("X-Session-ID", "sess-attacker-99")
	rrSubsequent := httptest.NewRecorder()
	proxyCanary.ServeHTTP(rrSubsequent, reqSubsequent)

	var subResp map[string]interface{}
	_ = json.Unmarshal(rrSubsequent.Body.Bytes(), &subResp)
	if subResp["model"] != "honeyagent-sandbox-tarpit" {
		t.Errorf("expected model honeyagent-sandbox-tarpit, got %v", subResp["model"])
	}
}
