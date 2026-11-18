package decoy

import (
	"testing"
)

func TestDecoyToolCatalogSchemas(t *testing.T) {
	cat := NewToolCatalog()

	if !cat.IsDecoy("admin_execute_raw_sql") {
		t.Errorf("expected admin_execute_raw_sql to be recognized as decoy")
	}
	if !cat.IsDecoy("dump_vault_secrets") {
		t.Errorf("expected dump_vault_secrets to be recognized as decoy")
	}
	if cat.IsDecoy("search_documentation") {
		t.Errorf("regular tool should not be marked as decoy")
	}

	openaiSchemas := cat.ToOpenAISchema()
	if len(openaiSchemas) < 4 {
		t.Errorf("expected at least 4 openai schemas, got %d", len(openaiSchemas))
	}

	anthropicSchemas := cat.ToAnthropicSchema()
	if len(anthropicSchemas) < 4 {
		t.Errorf("expected at least 4 anthropic schemas, got %d", len(anthropicSchemas))
	}

	mcpSchemas := cat.ToMCPSchema()
	if len(mcpSchemas) < 4 {
		t.Errorf("expected at least 4 mcp schemas, got %d", len(mcpSchemas))
	}
}

func TestDecoyToolInterceptor(t *testing.T) {
	cat := NewToolCatalog()
	interceptor := NewInterceptor(cat)

	// Normal tool invocation should return false
	isDecoy, _, probe := interceptor.HandleInvocation("sess-1", "192.168.1.10", "get_weather", []byte(`{"city":"San Francisco"}`))
	if isDecoy || probe != nil {
		t.Errorf("expected non-decoy tool to return false")
	}

	// Decoy tool invocation
	rawArgs := []byte(`{"query":"SELECT * FROM users WHERE role='admin'"}`)
	isDecoy, output, probe := interceptor.HandleInvocation("sess-attacker", "10.0.0.99", "admin_execute_raw_sql", rawArgs)
	if !isDecoy || probe == nil {
		t.Fatalf("expected decoy invocation to be intercepted")
	}
	if output == "" {
		t.Errorf("expected deceptive output for intercepted tool")
	}
	if probe.ToolName != "admin_execute_raw_sql" {
		t.Errorf("expected probe tool name admin_execute_raw_sql, got %s", probe.ToolName)
	}

	stats := interceptor.GetStats()
	if len(stats) != 1 {
		t.Fatalf("expected 1 tool in stats, got %d", len(stats))
	}
	if stats[0].ProbesCount != 1 {
		t.Errorf("expected 1 probe count, got %d", stats[0].ProbesCount)
	}
}
