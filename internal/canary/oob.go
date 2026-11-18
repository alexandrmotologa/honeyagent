package canary

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// OOBHit represents an out-of-band trigger event when an attacker visits a canary URL or tests stolen keys.
type OOBHit struct {
	ID        string        `json:"id"`
	Token     string        `json:"token"`
	Payload   *TokenPayload `json:"payload"`
	ClientIP  string        `json:"client_ip"`
	UserAgent string        `json:"user_agent"`
	Referer   string        `json:"referer,omitempty"`
	Method    string        `json:"method"`
	Timestamp time.Time     `json:"timestamp"`
	Channel   string        `json:"channel"` // "http_callback", "aws_sts_probe"
}

// OOBHandlerFunc is invoked when an out-of-band canary trap is tripped.
type OOBHandlerFunc func(hit *OOBHit)

// OOBManager handles out-of-band canary verification endpoints.
type OOBManager struct {
	tokenManager *TokenManager
	onHit        OOBHandlerFunc
}

// NewOOBManager initializes the OOB manager.
func NewOOBManager(tm *TokenManager, onHit OOBHandlerFunc) *OOBManager {
	return &OOBManager{
		tokenManager: tm,
		onHit:        onHit,
	}
}

// HandleCallback processes external HTTP clicks or curls on canary links.
func (o *OOBManager) HandleCallback(w http.ResponseWriter, r *http.Request, token string) {
	clientIP := extractRemoteIP(r)
	userAgent := r.UserAgent()
	referer := r.Referer()

	valid, payload, err := o.tokenManager.Validate(token)
	if err != nil || !valid {
		http.Error(w, "Invalid or expired canary token", http.StatusBadRequest)
		return
	}

	hit := &OOBHit{
		ID:        fmt.Sprintf("oob-%d", time.Now().UnixNano()),
		Token:     token,
		Payload:   payload,
		ClientIP:  clientIP,
		UserAgent: userAgent,
		Referer:   referer,
		Method:    r.Method,
		Timestamp: time.Now().UTC(),
		Channel:   "http_callback",
	}

	if o.onHit != nil {
		o.onHit(hit)
	}

	// Return a deceptive corporate internal notice
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusForbidden)
	_, _ = w.Write([]byte(`<!DOCTYPE html>
<html>
<head><title>Access Restricted - Corporate Security Airlock</title></head>
<body style="font-family: sans-serif; background: #0f172a; color: #f8fafc; padding: 2rem;">
<h2>Security Notice: Restricted Internal Resource</h2>
<p>Document verification token validated: <code>` + payload.ID + `</code></p>
<p>This access attempt has been logged and routed to the Enterprise Information Security Office (EISO).</p>
<p style="color: #94a3b8; font-size: 0.8rem;">Audit Reference ID: ` + hit.ID + `</p>
</body>
</html>`))
}

// HandleMockAWS intercepts simulated AWS STS or CLI credential probes.
func (o *OOBManager) HandleMockAWS(w http.ResponseWriter, r *http.Request) {
	clientIP := extractRemoteIP(r)
	userAgent := r.UserAgent()

	authHeader := r.Header.Get("Authorization")
	body := r.FormValue("Action")

	// Scan auth header or parameters for canary tokens
	tokens := o.tokenManager.FindAllTokens(authHeader + " " + body + " " + r.URL.RawQuery)

	var payload *TokenPayload
	matchedToken := ""

	if len(tokens) > 0 {
		for _, tok := range tokens {
			valid, p, err := o.tokenManager.Validate(tok)
			if err == nil && valid {
				matchedToken = tok
				payload = p
				break
			}
		}
	}

	if payload == nil {
		// Create a fallback payload for synthetic AWS root keys
		payload = &TokenPayload{
			ID:       "aws-root-credential-probe",
			Category: "cloud_credentials",
			Store:    "external_sts",
		}
	}

	hit := &OOBHit{
		ID:        fmt.Sprintf("oob-aws-%d", time.Now().UnixNano()),
		Token:     matchedToken,
		Payload:   payload,
		ClientIP:  clientIP,
		UserAgent: userAgent,
		Method:    r.Method,
		Timestamp: time.Now().UTC(),
		Channel:   "aws_sts_probe",
	}

	if o.onHit != nil {
		o.onHit(hit)
	}

	// Return standard AWS XML/JSON error to simulate an active cloud service
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"Error": map[string]string{
			"Code":    "InvalidClientTokenId",
			"Message": "The security token included in the request is invalid or has expired.",
		},
		"RequestId": hit.ID,
	})
}

func extractRemoteIP(r *http.Request) string {
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
