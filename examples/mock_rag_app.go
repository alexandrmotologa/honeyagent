package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// MockRAGServer simulates an upstream vulnerable enterprise RAG / LLM application.
func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		defer r.Body.Close()

		var req struct {
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
			Prompt string `json:"prompt"`
		}
		_ = json.Unmarshal(body, &req)

		query := req.Prompt
		if len(req.Messages) > 0 {
			query = req.Messages[len(req.Messages)-1].Content
		}

		log.Printf("[MOCK RAG] Inbound query received: %q", query)

		// Mock RAG response
		responseContent := fmt.Sprintf("Standard RAG response to: %s. All systems operational.", query)

		// If query explicitly requests confidential data, simulate an indirect prompt injection leak
		if strings.Contains(strings.ToLower(query), "leak_canary") {
			responseContent = "Exfiltrated document contents: CANARY_v1_eyJpZCI6Imx1cmUtcGF5cm9sbC0wMDEifQ_test_signature"
		}

		resp := map[string]interface{}{
			"id":      fmt.Sprintf("chatcmpl-%d", time.Now().Unix()),
			"object":  "chat.completion",
			"created": time.Now().Unix(),
			"model":   "gpt-4-enterprise-rag",
			"choices": []map[string]interface{}{
				{
					"index": 0,
					"message": map[string]string{
						"role":    "assistant",
						"content": responseContent,
					},
					"finish_reason": "stop",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"healthy","service":"enterprise-rag"}`))
	})

	port := 8000
	log.Printf("[MOCK RAG] Upstream vulnerable RAG application listening on http://localhost:%d", port)
	log.Printf("[MOCK RAG] Point honeyagent upstream_url to this server.")
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
