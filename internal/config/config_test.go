package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveAPIKeyFromSecrets_ByProvider(t *testing.T) {
	aiSecrets := map[string]any{
		"api_key": "default",
		"api_keys": map[string]any{
			"chaitin":           "k1",
			"doubao-ai":         "k2",
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

func TestLoad_AllowsMissingSecretsFile(t *testing.T) {
	dir := t.TempDir()
	appPath := filepath.Join(dir, "app.json")
	if err := os.WriteFile(appPath, []byte(`{"paths":{"state_dir":"state"}}`), 0o644); err != nil {
		t.Fatalf("write app.json: %v", err)
	}

	t.Setenv("YH_CONFIG", appPath)
	t.Setenv("YH_SECRETS", filepath.Join(dir, "missing.json"))

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got := cfg.PendingAuditsPath(); got != filepath.Join("state", "pending_audits.jsonl") {
		t.Fatalf("unexpected pending path: %q", got)
	}
}

func TestLoad_FailsOnInvalidSecretsJSON(t *testing.T) {
	dir := t.TempDir()
	appPath := filepath.Join(dir, "app.json")
	secretsPath := filepath.Join(dir, "secrets.json")
	if err := os.WriteFile(appPath, []byte(`{"paths":{"state_dir":"state"}}`), 0o644); err != nil {
		t.Fatalf("write app.json: %v", err)
	}
	if err := os.WriteFile(secretsPath, []byte(`{invalid json`), 0o644); err != nil {
		t.Fatalf("write secrets.json: %v", err)
	}

	t.Setenv("YH_CONFIG", appPath)
	t.Setenv("YH_SECRETS", secretsPath)

	if _, err := Load(); err == nil {
		t.Fatalf("expected error, got nil")
	}
}
