package server

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed all:static
var embeddedFiles embed.FS

// SPAHandler serves single-page application assets with index.html fallback for client-side routing.
func SPAHandler() http.Handler {
	distFS, err := fs.Sub(embeddedFiles, "static")
	if err != nil {
		return http.HandlerFunc(serveFallbackUI)
	}

	fileServer := http.FileServer(http.FS(distFS))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If request is for an API endpoint, do not serve SPA
		if strings.HasPrefix(r.URL.Path, "/api") {
			http.NotFound(w, r)
			return
		}

		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		// Check if file exists in embedded filesystem
		f, err := distFS.Open(path)
		if err == nil {
			_ = f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}

		// Fallback to index.html for client-side React routes
		indexFile, err := distFS.Open("index.html")
		if err == nil {
			_ = indexFile.Close()
			r.URL.Path = "/index.html"
			fileServer.ServeHTTP(w, r)
			return
		}

		// Standalone embedded fallback
		serveFallbackUI(w, r)
	})
}

func serveFallbackUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>honeyagent - Deception Studio</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #090d16; color: #f1f5f9; margin: 0; padding: 2rem; }
        .card { background: #131b2e; border: 1px solid #1e293b; border-radius: 8px; padding: 1.5rem; max-width: 800px; margin: 0 auto; box-shadow: 0 4px 6px -1px rgba(0,0,0,0.5); }
        h1 { color: #38bdf8; margin-top: 0; }
        .badge { background: #0284c7; color: white; padding: 4px 8px; border-radius: 4px; font-size: 12px; font-weight: bold; }
        button { background: #0284c7; color: white; border: none; padding: 8px 16px; border-radius: 4px; cursor: pointer; font-weight: 500; margin-right: 8px; }
        button:hover { background: #0369a1; }
        pre { background: #050811; padding: 1rem; border-radius: 6px; overflow-x: auto; color: #a5f3fc; }
    </style>
</head>
<body>
    <div class="card">
        <h1>honeyagent <span class="badge">Airlock Active</span></h1>
        <p>Enterprise AI Deception Mesh & Canary Tripwire Gateway.</p>
        <div style="margin: 1.5rem 0;">
            <button onclick="runSimulation('canary_exfiltration')">Simulate Canary Retrieval</button>
            <button onclick="runSimulation('decoy_tool')">Simulate Decoy Tool Call</button>
            <button onclick="fetchStats()">Refresh Stats</button>
        </div>
        <pre id="output">Loading telemetry...</pre>
    </div>
    <script>
        async function fetchStats() {
            const res = await fetch('/api/stats');
            const data = await res.json();
            document.getElementById('output').textContent = JSON.stringify(data, null, 2);
        }
        async function runSimulation(type) {
            document.getElementById('output').textContent = 'Executing adversarial simulation...';
            const res = await fetch('/api/simulate', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ type: type })
            });
            const data = await res.json();
            document.getElementById('output').textContent = JSON.stringify(data, null, 2);
        }
        fetchStats();
    </script>
</body>
</html>`))
}
