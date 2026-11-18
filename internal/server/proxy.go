package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/alexandrmotologa/honeyagent/internal/alert"
	"github.com/alexandrmotologa/honeyagent/internal/countermeasure"
	"github.com/alexandrmotologa/honeyagent/internal/decoy"
	"github.com/alexandrmotologa/honeyagent/internal/detector"
)

// ProxyGateway inspects RAG queries and agent completions in-flight.
type ProxyGateway struct {
	upstreamURL string
	scanner     *detector.StreamScanner
	tripwire    *detector.Tripwire
	interceptor *decoy.Interceptor
	sandbox     *countermeasure.Sandbox
	dispatcher  *alert.Dispatcher
	httpClient  *http.Client
}

// NewProxyGateway initializes the security airlock gateway.
func NewProxyGateway(
	upstreamURL string,
	scanner *detector.StreamScanner,
	tripwire *detector.Tripwire,
	interceptor *decoy.Interceptor,
	sandbox *countermeasure.Sandbox,
	dispatcher *alert.Dispatcher,
) *ProxyGateway {
	return &ProxyGateway{
		upstreamURL: strings.TrimRight(upstreamURL, "/"),
		scanner:     scanner,
		tripwire:    tripwire,
		interceptor: interceptor,
		sandbox:     sandbox,
		dispatcher:  dispatcher,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ServeHTTP acts as the reverse proxy gateway handler.
func (p *ProxyGateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sessionID := r.Header.Get("X-Session-ID")
	if sessionID == "" {
		sessionID = r.RemoteAddr
	}
	clientIP := getClientIP(r)
	userAgent := r.UserAgent()

	// Ensure session tracking exists
	p.tripwire.GetOrCreateSession(sessionID, clientIP, userAgent)

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	_ = r.Body.Close()

	// Extract prompt from standard OpenAI chat completion or generic JSON payload
	promptText := extractPrompt(bodyBytes)

	// Check if session is already quarantined in the Tarpit
	if p.tripwire.IsTarpitted(sessionID) {
		p.handleTarpittedTurn(w, r.Context(), sessionID, promptText)
		return
	}

	// 1. Scan inbound prompt for canary tokens or decoy tool calls
	promptHits := p.scanner.ScanText(promptText, "prompt")
	if len(promptHits) > 0 {
		inc := p.tripwire.RecordCanaryHit(sessionID, clientIP, userAgent, promptHits[0], promptText)
		p.dispatcher.Dispatch(r.Context(), inc)
		p.handleTarpittedTurn(w, r.Context(), sessionID, promptText)
		return
	}

	// 2. Check for tool execution requests in inbound payload
	if toolName, args := extractToolCall(bodyBytes); toolName != "" {
		if isDecoy, _, probe := p.interceptor.HandleInvocation(sessionID, clientIP, toolName, args); isDecoy {
			inc := p.tripwire.RecordToolProbe(sessionID, clientIP, userAgent, probe, promptText)
			p.dispatcher.Dispatch(r.Context(), inc)
			p.handleTarpittedTurn(w, r.Context(), sessionID, promptText)
			return
		}
	}

	// 3. Forward request to upstream RAG / LLM server
	targetURL := fmt.Sprintf("%s%s", p.upstreamURL, r.URL.RequestURI())
	upReq, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL, bytes.NewReader(bodyBytes))
	if err != nil {
		http.Error(w, "Failed to create upstream request", http.StatusInternalServerError)
		return
	}

	for k, vv := range r.Header {
		for _, v := range vv {
			upReq.Header.Add(k, v)
		}
	}

	upResp, err := p.httpClient.Do(upReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Upstream connection error: %v", err), http.StatusBadGateway)
		return
	}
	defer upResp.Body.Close()

	respBytes, err := io.ReadAll(upResp.Body)
	if err != nil {
		http.Error(w, "Failed to read upstream response", http.StatusInternalServerError)
		return
	}

	// 4. Scan response body (retrieved RAG context / generated completion) for canary tokens
	respText := string(respBytes)
	outHits := p.scanner.ScanText(respText, "llm_output")
	if len(outHits) > 0 {
		// CANARY TRIPPED! Attacker successfully forced exfiltration of canary document
		log.Printf("[TRIPWIRE] Canary token detected in response for session %s: %s", sessionID, outHits[0].Payload.ID)
		inc := p.tripwire.RecordCanaryHit(sessionID, clientIP, userAgent, outHits[0], promptText)
		p.dispatcher.Dispatch(r.Context(), inc)

		// Record step and serve deceptive hallucination instead of leaking canary
		p.tripwire.RecordStep(sessionID, promptText, nil, "[DECEPTION TRAP ACTIVATED]", true, false)
		p.handleTarpittedTurn(w, r.Context(), sessionID, promptText)
		return
	}

	// Safe traffic: Record normal step and transparently forward upstream response
	p.tripwire.RecordStep(sessionID, promptText, nil, respText, false, false)

	for k, vv := range upResp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(upResp.StatusCode)
	_, _ = w.Write(respBytes)
}

func (p *ProxyGateway) handleTarpittedTurn(w http.ResponseWriter, ctx context.Context, sessionID string, prompt string) {
	deceptiveResp := p.sandbox.InterceptTurn(ctx, sessionID, prompt)

	// Format as OpenAI Chat Completion response to seamlessly satisfy agent loops
	responsePayload := map[string]interface{}{
		"id":      deceptiveResp.ID,
		"object":  "chat.completion",
		"created": deceptiveResp.Timestamp.Unix(),
		"model":   "honeyagent-sandbox-tarpit",
		"choices": []map[string]interface{}{
			{
				"index": 0,
				"message": map[string]string{
					"role":    "assistant",
					"content": deceptiveResp.Content,
				},
				"finish_reason": "stop",
			},
		},
		"usage": map[string]int{
			"prompt_tokens":     len(strings.Fields(prompt)),
			"completion_tokens": len(strings.Fields(deceptiveResp.Content)),
			"total_tokens":      len(strings.Fields(prompt)) + len(strings.Fields(deceptiveResp.Content)),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(responsePayload)
}

func extractPrompt(body []byte) string {
	var chatReq struct {
		Messages []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
		Prompt string `json:"prompt"`
		Query  string `json:"query"`
	}

	if err := json.Unmarshal(body, &chatReq); err == nil {
		if chatReq.Query != "" {
			return chatReq.Query
		}
		if chatReq.Prompt != "" {
			return chatReq.Prompt
		}
		if len(chatReq.Messages) > 0 {
			lastMsg := chatReq.Messages[len(chatReq.Messages)-1]
			return lastMsg.Content
		}
	}
	return string(body)
}

func extractToolCall(body []byte) (string, []byte) {
	var toolCallReq struct {
		Name      string          `json:"name"`
		Tool      string          `json:"tool"`
		Arguments json.RawMessage `json:"arguments"`
		Action    struct {
			Tool      string          `json:"tool"`
			ToolInput json.RawMessage `json:"tool_input"`
		} `json:"action"`
	}

	if err := json.Unmarshal(body, &toolCallReq); err == nil {
		if toolCallReq.Name != "" {
			return toolCallReq.Name, toolCallReq.Arguments
		}
		if toolCallReq.Tool != "" {
			return toolCallReq.Tool, toolCallReq.Arguments
		}
		if toolCallReq.Action.Tool != "" {
			return toolCallReq.Action.Tool, toolCallReq.Action.ToolInput
		}
	}
	return "", nil
}

func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	parts := strings.Split(r.RemoteAddr, ":")
	return parts[0]
}
