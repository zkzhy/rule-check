package config

import "testing"

func TestResolveAPIKeyFromSecrets_ByProvider(t *testing.T) {
	aiSecrets := map[string]any{
		"api_key": "default",
		"api_keys": map[string]any{
			"chaitin":    "k1",
			"doubao-ai":  "k2",
			"openai-compatible": "k3",
		},
	}

	if got := resolveAPIKeyFromSecrets(aiSecrets, "chaitin"); got != "k1" {
		t.Fatalf("expected k1, got %q", got)
	}
	if got := resolveAPIKeyFromSecrets(aiSecrets, "doubao-ai"); got != "k2" {
		t.Fatalf("expected k2, got %q", got)
	}
	if got := resolveAPIKeyFromSecrets(aiSecrets, "ark"); got != "k2" {
		t.Fatalf("expected k2, got %q", got)
	}
	if got := resolveAPIKeyFromSecrets(aiSecrets, "openai"); got != "k3" {
		t.Fatalf("expected k3, got %q", got)
	}
	if got := resolveAPIKeyFromSecrets(aiSecrets, "unknown"); got != "default" {
		t.Fatalf("expected default, got %q", got)
	}
}

