package adr

import (
	"os"
	"path/filepath"
	"testing"
)

const testADR1 = `# ADR-0001: Test Decision

- **Date**: 2026-01-01
- **Status**: Accepted
- **Domain(s)**: core, context-system
- **Supersedes**: n/a
- **Superseded-by**: n/a

## Decision
Some decision.
`

const testADR2 = `# ADR-0002: Another Decision

- **Date**: 2026-01-02
- **Status**: Accepted
- **Domain(s)**: workflow
- **Supersedes**: n/a
- **Superseded-by**: n/a

## Decision
Another.
`

const testADR3 = `# ADR-0003: Superseded One

- **Date**: 2026-01-03
- **Status**: Superseded
- **Domain(s)**: core
- **Supersedes**: n/a
- **Superseded-by**: ADR-0005

## Decision
Old.
`

func setupTestADRs(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	adrDir := filepath.Join(root, "docs", "adr")
	os.MkdirAll(adrDir, 0o755)
	os.WriteFile(filepath.Join(adrDir, "ADR-0001.md"), []byte(testADR1), 0o644)
	os.WriteFile(filepath.Join(adrDir, "ADR-0002.md"), []byte(testADR2), 0o644)
	os.WriteFile(filepath.Join(adrDir, "ADR-0003.md"), []byte(testADR3), 0o644)
	return root
}

func TestParseADR(t *testing.T) {
	root := setupTestADRs(t)
	path := filepath.Join(root, "docs", "adr", "ADR-0001.md")

	a, err := ParseADR(path)
	if err != nil {
		t.Fatalf("ParseADR: %v", err)
	}

	if a.ID != "ADR-0001" {
		t.Errorf("ID = %q, want ADR-0001", a.ID)
	}
	if a.Title != "Test Decision" {
		t.Errorf("Title = %q, want %q", a.Title, "Test Decision")
	}
	if a.Date != "2026-01-01" {
		t.Errorf("Date = %q, want 2026-01-01", a.Date)
	}
	if a.Status != "Accepted" {
		t.Errorf("Status = %q, want Accepted", a.Status)
	}
	if len(a.Domains) != 2 || a.Domains[0] != "core" || a.Domains[1] != "context-system" {
		t.Errorf("Domains = %v, want [core context-system]", a.Domains)
	}
	if a.Supersedes != "" {
		t.Errorf("Supersedes = %q, want empty (n/a)", a.Supersedes)
	}
	if a.SupersededBy != "" {
		t.Errorf("SupersededBy = %q, want empty (n/a)", a.SupersededBy)
	}
}

func TestParseADR_SupersededBy(t *testing.T) {
	root := setupTestADRs(t)
	path := filepath.Join(root, "docs", "adr", "ADR-0003.md")

	a, err := ParseADR(path)
	if err != nil {
		t.Fatalf("ParseADR: %v", err)
	}

	if a.Status != "Superseded" {
		t.Errorf("Status = %q, want Superseded", a.Status)
	}
	if a.SupersededBy != "ADR-0005" {
		t.Errorf("SupersededBy = %q, want ADR-0005", a.SupersededBy)
	}
}

func TestScanADRs_Sorted(t *testing.T) {
	root := setupTestADRs(t)

	adrs, err := ScanADRs(root)
	if err != nil {
		t.Fatalf("ScanADRs: %v", err)
	}

	if len(adrs) != 3 {
		t.Fatalf("got %d ADRs, want 3", len(adrs))
	}

	// Verify sorted by ID
	for i := 1; i < len(adrs); i++ {
		if adrs[i].ID < adrs[i-1].ID {
			t.Errorf("ADRs not sorted: %s before %s", adrs[i-1].ID, adrs[i].ID)
		}
	}

	// Verify all fields populated
	if adrs[0].Title != "Test Decision" {
		t.Errorf("adrs[0].Title = %q, want %q", adrs[0].Title, "Test Decision")
	}
	if adrs[0].Date != "2026-01-01" {
		t.Errorf("adrs[0].Date = %q, want 2026-01-01", adrs[0].Date)
	}
}

func TestScanADRs_EmptyDir(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "docs", "adr"), 0o755)

	adrs, err := ScanADRs(root)
	if err != nil {
		t.Fatalf("ScanADRs: %v", err)
	}
	if len(adrs) != 0 {
		t.Errorf("got %d ADRs, want 0", len(adrs))
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

func TestNextID(t *testing.T) {
	root := setupTestADRs(t)

	id, err := NextID(root)
	if err != nil {
		t.Fatalf("NextID: %v", err)
	}
	if id != "0004" {
		t.Errorf("NextID = %q, want 0004", id)
	}
}

func TestNextID_Empty(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "docs", "adr"), 0o755)

	id, err := NextID(root)
	if err != nil {
		t.Fatalf("NextID: %v", err)
	}
	if id != "0001" {
		t.Errorf("NextID = %q, want 0001", id)
	}
}
