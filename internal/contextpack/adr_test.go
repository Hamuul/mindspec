package contextpack

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanADRs(t *testing.T) {
	root := t.TempDir()
	adrDir := filepath.Join(root, "docs", "adr")
	os.MkdirAll(adrDir, 0o755)

	adr1 := `# ADR-0001: Test Decision

- **Date**: 2026-01-01
- **Status**: Accepted
- **Domain(s)**: core, context-system

## Decision
Some decision.
`
	adr2 := `# ADR-0002: Another Decision

- **Date**: 2026-01-02
- **Status**: Accepted
- **Domain(s)**: workflow

## Decision
Another.
`
	adr3 := `# ADR-0003: Superseded

- **Date**: 2026-01-03
- **Status**: Superseded
- **Domain(s)**: core

## Decision
Old.
`

	os.WriteFile(filepath.Join(adrDir, "ADR-0001.md"), []byte(adr1), 0o644)
	os.WriteFile(filepath.Join(adrDir, "ADR-0002.md"), []byte(adr2), 0o644)
	os.WriteFile(filepath.Join(adrDir, "ADR-0003.md"), []byte(adr3), 0o644)

	adrs, err := ScanADRs(root)
	if err != nil {
		t.Fatalf("ScanADRs: %v", err)
	}

	if len(adrs) != 3 {
		t.Fatalf("got %d ADRs, want 3", len(adrs))
	}
}

func TestFilterADRs(t *testing.T) {
	adrs := []ADR{
		{ID: "ADR-0001", Status: "Accepted", Domains: []string{"core", "context-system"}},
		{ID: "ADR-0002", Status: "Accepted", Domains: []string{"workflow"}},
		{ID: "ADR-0003", Status: "Superseded", Domains: []string{"core"}},
	}

	filtered := FilterADRs(adrs, []string{"context-system"})
	if len(filtered) != 1 {
		t.Fatalf("got %d filtered ADRs, want 1", len(filtered))
	}
	if filtered[0].ID != "ADR-0001" {
		t.Errorf("filtered[0].ID = %q, want ADR-0001", filtered[0].ID)
	}
}

func TestFilterADRs_MultiDomain(t *testing.T) {
	adrs := []ADR{
		{ID: "ADR-0001", Status: "Accepted", Domains: []string{"core", "context-system"}},
		{ID: "ADR-0002", Status: "Accepted", Domains: []string{"workflow", "context-system"}},
	}

	filtered := FilterADRs(adrs, []string{"context-system", "workflow"})
	if len(filtered) != 2 {
		t.Fatalf("got %d filtered ADRs, want 2", len(filtered))
	}
}
