package adr

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSupersede(t *testing.T) {
	root := setupTestADRs(t)

	err := Supersede(root, "ADR-0001", "ADR-0005")
	if err != nil {
		t.Fatalf("Supersede: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(root, "docs", "adr", "ADR-0001.md"))
	if !strings.Contains(string(data), "**Superseded-by**: ADR-0005") {
		t.Errorf("expected Superseded-by update, got:\n%s", string(data))
	}
}

func TestCopyDomains(t *testing.T) {
	root := setupTestADRs(t)

	domains, err := CopyDomains(root, "ADR-0001")
	if err != nil {
		t.Fatalf("CopyDomains: %v", err)
	}

	if len(domains) != 2 || domains[0] != "core" || domains[1] != "context-system" {
		t.Errorf("domains = %v, want [core context-system]", domains)
	}
}

func TestSupersede_MissingADR(t *testing.T) {
	root := setupTestADRs(t)

	err := Supersede(root, "ADR-9999", "ADR-0005")
	if err == nil {
		t.Error("expected error for missing ADR")
	}
}
