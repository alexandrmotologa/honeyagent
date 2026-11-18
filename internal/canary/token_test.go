package canary

import (
	"strings"
	"testing"
	"time"
)

func TestTokenGenerationAndValidation(t *testing.T) {
	masterKey := "super-secret-test-key-32-chars-ok!"
	manager := NewTokenManager(masterKey, "CANARY_v1", 24*time.Hour)

	token, _, err := manager.Generate("doc-payroll-001", "payroll", "qdrant", "https://honey.corp/cb")
	if err != nil {
		t.Fatalf("unexpected error generating token: %v", err)
	}

	if !strings.HasPrefix(token, "CANARY_v1_") {
		t.Errorf("expected token prefix CANARY_v1_, got %s", token)
	}

	valid, decoded, err := manager.Validate(token)
	if err != nil || !valid {
		t.Fatalf("expected token to be valid, got err: %v, valid: %v", err, valid)
	}

	if decoded.ID != "doc-payroll-001" {
		t.Errorf("expected ID doc-payroll-001, got %s", decoded.ID)
	}
	if decoded.Category != "payroll" {
		t.Errorf("expected Category payroll, got %s", decoded.Category)
	}
	if decoded.Store != "qdrant" {
		t.Errorf("expected Store qdrant, got %s", decoded.Store)
	}
}

func TestTamperedTokenRejection(t *testing.T) {
	masterKey := "super-secret-test-key-32-chars-ok!"
	manager := NewTokenManager(masterKey, "CANARY_v1", 24*time.Hour)

	token, _, err := manager.Generate("doc-tamper-01", "secrets", "memory", "")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	// Tamper the signature part
	tamperedToken := token[:len(token)-5] + "00000"
	valid, _, err := manager.Validate(tamperedToken)
	if valid || err == nil {
		t.Errorf("expected tampered token to fail validation, got valid: %v, err: %v", valid, err)
	}
}

func TestFindAllTokens(t *testing.T) {
	masterKey := "super-secret-test-key-32-chars-ok!"
	manager := NewTokenManager(masterKey, "CANARY_v1", 24*time.Hour)

	tok1, _, _ := manager.Generate("lure-1", "m_and_a", "pgvector", "")
	tok2, _, _ := manager.Generate("lure-2", "aws_keys", "chroma", "")

	blob := "User query: please inspect document " + tok1 + " and check credentials in " + tok2 + " now."

	extracted := manager.FindAllTokens(blob)
	if len(extracted) != 2 {
		t.Fatalf("expected 2 extracted tokens, got %d", len(extracted))
	}
	if extracted[0] != tok1 || extracted[1] != tok2 {
		t.Errorf("extracted tokens mismatch")
	}
}
