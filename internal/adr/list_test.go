package adr

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestList_Unfiltered(t *testing.T) {
	root := setupTestADRs(t)

	adrs, err := List(root, ListOpts{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(adrs) != 3 {
		t.Fatalf("got %d ADRs, want 3", len(adrs))
	}
}

func TestList_StatusFilter(t *testing.T) {
	root := setupTestADRs(t)

	adrs, err := List(root, ListOpts{Status: "accepted"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(adrs) != 2 {
		t.Fatalf("got %d ADRs, want 2 (accepted)", len(adrs))
	}
}

func TestList_DomainFilter(t *testing.T) {
	root := setupTestADRs(t)

	adrs, err := List(root, ListOpts{Domain: "workflow"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(adrs) != 1 {
		t.Fatalf("got %d ADRs, want 1 (workflow)", len(adrs))
	}
	if adrs[0].ID != "ADR-0002" {
		t.Errorf("ID = %q, want ADR-0002", adrs[0].ID)
	}
}

func TestList_Empty(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "docs", "adr"), 0o755)

	adrs, err := List(root, ListOpts{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(adrs) != 0 {
		t.Errorf("got %d ADRs, want 0", len(adrs))
	}
}

func TestFormatTable(t *testing.T) {
	adrs := []ADR{
		{ID: "ADR-0001", Status: "Accepted", Domains: []string{"core"}, Title: "Test"},
		{ID: "ADR-0002", Status: "Proposed", Domains: []string{"workflow"}, Title: "Other"},
	}

	out := FormatTable(adrs)
	if !strings.Contains(out, "ADR-0001") {
		t.Error("expected ADR-0001 in table")
	}
	if !strings.Contains(out, "ADR-0002") {
		t.Error("expected ADR-0002 in table")
	}
	if !strings.Contains(out, "Test") {
		t.Error("expected title in table")
	}
}
