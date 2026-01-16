package fetch

import "testing"

func TestResolveURL_OverridesBasePath(t *testing.T) {
	u, err := resolveURL("https://example.com/login", "/api/login")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if u != "https://example.com/api/login" {
		t.Fatalf("unexpected url: %s", u)
	}
}

func TestNormalizePath(t *testing.T) {
	if got := normalizePath("api/lines/operation"); got != "/api/lines/operation" {
		t.Fatalf("unexpected: %s", got)
	}
	if got := normalizePath("/api/lines/operation"); got != "/api/lines/operation" {
		t.Fatalf("unexpected: %s", got)
	}
}

