package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/alexandrmotologa/honeyagent/internal/detector"
)

// SSEBroker streams real-time threat events to connected React 19 dashboards.
type SSEBroker struct {
	tripwire *detector.Tripwire
}

// NewSSEBroker creates an SSE broker.
func NewSSEBroker(tripwire *detector.Tripwire) *SSEBroker {
	return &SSEBroker{tripwire: tripwire}
}

// ServeHTTP handles Server-Sent Events subscriptions.
func (b *SSEBroker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	eventCh, unsubscribe := b.tripwire.Subscribe()
	defer unsubscribe()

	// Initial connection confirmation
	_, _ = fmt.Fprintf(w, "event: connected\ndata: {\"status\":\"connected\",\"time\":\"%s\"}\n\n", time.Now().UTC().Format(time.RFC3339))
	flusher.Flush()

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case inc, ok := <-eventCh:
			if !ok {
				return
			}
			data, err := json.Marshal(inc)
			if err == nil {
				_, _ = fmt.Fprintf(w, "event: incident\ndata: %s\n\n", string(data))
				flusher.Flush()
			}
		case <-ticker.C:
			// Heartbeat
			_, _ = fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()
		}
	}
}
