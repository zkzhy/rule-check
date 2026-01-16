package orchestrator

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"audit-workflow/internal/config"
)

func TestBuildTrimmedContext_RespectsBudgets(t *testing.T) {
	cfg := &config.RootConfig{
		AI: config.AIConfig{
			Context: config.AIContextConfig{
				TotalMaxRunes:       60,
				NameMaxRunes:        10,
				DescriptionMaxRunes: 20,
				POCMaxRunes:         20,
				ReqMaxRunes:         20,
				RespMaxRunes:        20,
			},
		},
	}

	data := map[string]any{
		"name":             strings.Repeat("名", 50),
		"description":      strings.Repeat("描", 50),
		"xray_poc_content": strings.Repeat("证", 50),
		"req_pkg":          strings.Repeat("请", 50),
		"resp_pkg":         strings.Repeat("响", 50),
	}

	out := buildTrimmedContext(cfg, data)
	if out == "" {
		t.Fatalf("expected non-empty output")
	}
	if got := len([]rune(out)); got > cfg.AI.Context.TotalMaxRunes {
		t.Fatalf("expected <= %d runes, got %d: %q", cfg.AI.Context.TotalMaxRunes, got, out)
	}

	if !strings.HasPrefix(out, "漏洞名称：") {
		t.Fatalf("expected to start with name section, got: %q", out)
	}
}

func TestLoadPendingRecords_LongLines(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "pending.jsonl")

	long := strings.Repeat("x", 200*1024)
	line1 := `{"id":1,"data":{"name":"a"}}` + "\n"
	line2 := `{"id":2,"data":{"description":"` + long + `"}}` + "\n"
	if err := os.WriteFile(p, []byte(line1+line2), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	items, err := loadPendingRecords(p)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
}

func TestLLMLimiter_WaitHonorsContextCancel(t *testing.T) {
	limiter := newLLMLimiter(1)
	defer limiter.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	if err := limiter.Wait(ctx); err == nil {
		t.Fatalf("expected context cancellation error")
	}
}

func TestLLMLimiter_WaitEventuallyReturns(t *testing.T) {
	limiter := newLLMLimiter(1000)
	defer limiter.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	if err := limiter.Wait(ctx); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}
