package canary

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var (
	ErrInvalidTokenFormat = errors.New("invalid canary token format")
	ErrSignatureMismatch  = errors.New("canary token signature mismatch")
	ErrTokenExpired       = errors.New("canary token has expired")
)

// TokenPayload carries metadata embedded in a canary token.
type TokenPayload struct {
	ID          string `json:"id"`
	Category    string `json:"cat"`
	Store       string `json:"st"`
	CreatedAt   int64  `json:"ts"`
	ExpiresAt   int64  `json:"exp,omitempty"`
	Nonce       string `json:"rnd"`
	CallbackURL string `json:"cb,omitempty"`
}

// TokenManager handles generation, signing, and verification of canary tokens.
type TokenManager struct {
	masterKey []byte
	prefix    string
	ttl       time.Duration
}

// NewTokenManager initializes a manager with the given master key and options.
func NewTokenManager(masterKey string, prefix string, ttl time.Duration) *TokenManager {
	if prefix == "" {
		prefix = "CANARY_v1"
	}
	if ttl <= 0 {
		ttl = 30 * 24 * time.Hour
	}
	return &TokenManager{
		masterKey: []byte(masterKey),
		prefix:    prefix,
		ttl:       ttl,
	}
}

// Generate creates a cryptographically signed canary token string.
func (m *TokenManager) Generate(id string, category string, store string, callbackURL string) (string, *TokenPayload, error) {
	nonceBytes := make([]byte, 8)
	if _, err := rand.Read(nonceBytes); err != nil {
		return "", nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	now := time.Now().UTC()
	payload := &TokenPayload{
		ID:          id,
		Category:    category,
		Store:       store,
		CreatedAt:   now.Unix(),
		ExpiresAt:   now.Add(m.ttl).Unix(),
		Nonce:       hex.EncodeToString(nonceBytes),
		CallbackURL: callbackURL,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal token payload: %w", err)
	}

	encodedPayload := base64.RawURLEncoding.EncodeToString(payloadJSON)
	sig := m.computeHMAC(encodedPayload)

	token := fmt.Sprintf("%s_%s_%s", m.prefix, encodedPayload, sig)
	return token, payload, nil
}

// Validate verifies the token signature and returns the payload if genuine.
func (m *TokenManager) Validate(token string) (bool, *TokenPayload, error) {
	parts := strings.Split(token, "_")
	if len(parts) < 3 {
		return false, nil, ErrInvalidTokenFormat
	}

	// Reconstruct prefix, encoded payload, and signature
	// Format is: <prefix>_<encodedPayload>_<signature>
	sig := parts[len(parts)-1]
	encodedPayload := parts[len(parts)-2]
	prefix := strings.Join(parts[:len(parts)-2], "_")

	if prefix != m.prefix {
		return false, nil, ErrInvalidTokenFormat
	}

	expectedSig := m.computeHMAC(encodedPayload)
	expectedSigBytes, err1 := hex.DecodeString(expectedSig)
	actualSigBytes, err2 := hex.DecodeString(sig)
	if err1 != nil || err2 != nil || !hmac.Equal(expectedSigBytes, actualSigBytes) {
		return false, nil, ErrSignatureMismatch
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(encodedPayload)
	if err != nil {
		return false, nil, fmt.Errorf("failed to decode payload: %w", err)
	}

	var payload TokenPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return false, nil, fmt.Errorf("failed to parse payload json: %w", err)
	}

	if payload.ExpiresAt > 0 && time.Now().UTC().Unix() > payload.ExpiresAt {
		return false, &payload, ErrTokenExpired
	}

	return true, &payload, nil
}

// computeHMAC generates a hex-encoded HMAC-SHA256 signature for the data.
func (m *TokenManager) computeHMAC(data string) string {
	mac := hmac.New(sha256.New, m.masterKey)
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}

// FindAllTokens extracts potential canary tokens from raw input text.
func (m *TokenManager) FindAllTokens(text string) []string {
	pattern := regexp.MustCompile(fmt.Sprintf(`%s_[A-Za-z0-9_-]+_[0-9a-fA-F]{64}`, regexp.QuoteMeta(m.prefix)))
	return pattern.FindAllString(text, -1)
}
