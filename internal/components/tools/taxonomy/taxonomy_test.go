package taxonomy

import (
	"os"
	"strings"
	"testing"
)

func TestLoadAndLookup(t *testing.T) {
	pathsToCheck := []string{
		"ATT&CK.csv",
		"../ATT&CK.csv",
	}

	var csvPath string
	for _, p := range pathsToCheck {
		if _, err := os.Stat(p); err == nil {
			csvPath = p
			break
		}
	}
	if csvPath == "" {
		t.Skip("ATT&CK.csv not found, skipping test")
	}

	err := Load(csvPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	tid, teid, subid, found := LookupIDs("侦察", "主动扫描", "扫描 IP 块")
	if !found {
		t.Errorf("Lookup failed for 侦察/主动扫描/扫描 IP 块")
	}
	if tid != 1 || teid != 1 || subid != 2 {
		t.Errorf("ID mismatch: got %d,%d,%d; want 1,1,2", tid, teid, subid)
	}

	id, ok := LookupTacticID("侦察")
	if !ok || id != 1 {
		t.Errorf("LookupTacticID failed for 侦察: got %d, %v", id, ok)
	}
}

func TestCandidatesGeneration(t *testing.T) {
	pathsToCheck := []string{
		"ATT&CK.csv",
		"../ATT&CK.csv",
	}

	var csvPath string
	for _, p := range pathsToCheck {
		if _, err := os.Stat(p); err == nil {
			csvPath = p
			break
		}
	}
	if csvPath == "" {
		t.Skip("ATT&CK.csv not found, skipping test")
	}

	if err := Load(csvPath); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	tactics := ListTactics()
	if len(tactics) == 0 {
		t.Fatalf("expected non-empty tactics list")
	}
	foundRecon := false
	for _, tName := range tactics {
		if tName == "侦察" {
			foundRecon = true
			break
		}
	}
	if !foundRecon {
		t.Fatalf("expected tactics to include 侦察, got: %v", tactics)
	}

	cands := GenerateTechniqueCandidates("侦察", "扫描 IP 块 漏洞扫描", 5, 5)
	if len(cands) == 0 {
		t.Fatalf("expected non-empty candidates")
	}
	gotText := FormatTechniqueCandidates("侦察", cands, 500)
	if !strings.Contains(gotText, "主动扫描") {
		t.Fatalf("expected candidates to include 主动扫描, got: %q", gotText)
	}
}
