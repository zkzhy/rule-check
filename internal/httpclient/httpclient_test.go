package httpclient

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDoJSON_HTMLResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body>login</body></html>"))
	}))
	defer srv.Close()

	cl := New(true, 5)
	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	var out map[string]any
	code, err := cl.DoJSON(req, &out)
	if code != 200 {
		t.Fatalf("expected 200, got %d", code)
	}
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "expected JSON but got HTML") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDoJSON_JSONResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	cl := New(true, 5)
	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	var out map[string]any
	code, err := cl.DoJSON(req, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != 200 {
		t.Fatalf("expected 200, got %d", code)
	}
	if v, ok := out["ok"].(bool); !ok || !v {
		t.Fatalf("unexpected decoded output: %#v", out)
	}
}

func TestDoJSON_LargeJSONResponse(t *testing.T) {
	blob := strings.Repeat("a", 2<<20)
	var payload bytes.Buffer
	payload.WriteString(`{"ok":true,"blob":"`)
	payload.WriteString(blob)
	payload.WriteString(`"}`)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(payload.Bytes())
	}))
	defer srv.Close()

	cl := New(true, 5)
	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}

	var out map[string]any
	code, err := cl.DoJSON(req, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != 200 {
		t.Fatalf("expected 200, got %d", code)
	}
	if v, ok := out["blob"].(string); !ok || len(v) != len(blob) {
		t.Fatalf("unexpected decoded output: blob_len=%v", len(v))
	}
}
