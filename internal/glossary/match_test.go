package glossary

import (
	"testing"
)

var matchEntries = []Entry{
	{Term: "ADR", Target: "docs/core/ARCHITECTURE.md#adr-lifecycle"},
	{Term: "Bead", Target: "docs/core/ARCHITECTURE.md#beads"},
	{Term: "Context Pack", Target: "docs/core/ARCHITECTURE.md#context-system"},
	{Term: "Context System", Target: "docs/core/ARCHITECTURE.md#context-system"},
	{Term: "Spec Mode", Target: "docs/core/MODES.md#spec-mode"},
	{Term: "Implementation Mode", Target: "docs/core/MODES.md#implementation-mode"},
}

func TestMatch_Exact(t *testing.T) {
	results := Match(matchEntries, "Bead")
	if len(results) != 1 {
		t.Fatalf("expected 1 match, got %d", len(results))
	}
	if results[0].Term != "Bead" {
		t.Errorf("expected 'Bead', got %q", results[0].Term)
	}
}

func TestMatch_CaseInsensitive(t *testing.T) {
	results := Match(matchEntries, "CONTEXT PACK")
	found := false
	for _, r := range results {
		if r.Term == "Context Pack" {
			found = true
		}
	}
	if !found {
		t.Error("expected case-insensitive match for 'Context Pack'")
	}
}

func TestMatch_Multiple(t *testing.T) {
	results := Match(matchEntries, "context pack and bead")
	terms := make(map[string]bool)
	for _, r := range results {
		terms[r.Term] = true
	}
	if !terms["Context Pack"] {
		t.Error("expected 'Context Pack' in results")
	}
	if !terms["Bead"] {
		t.Error("expected 'Bead' in results")
	}
}

func TestMatch_LongestFirst(t *testing.T) {
	results := Match(matchEntries, "implementation mode")
	if len(results) == 0 {
		t.Fatal("expected at least one match")
	}
	// "Implementation Mode" (19 chars) should come before "Bead" etc.
	if results[0].Term != "Implementation Mode" {
		t.Errorf("expected longest match first, got %q", results[0].Term)
	}
}

func TestMatch_NoMatch(t *testing.T) {
	results := Match(matchEntries, "nothing relevant here")
	if len(results) != 0 {
		t.Errorf("expected no matches, got %d", len(results))
	}
}

func TestMatch_SpecModeInSentence(t *testing.T) {
	results := Match(matchEntries, "spec mode approval")
	found := false
	for _, r := range results {
		if r.Term == "Spec Mode" {
			found = true
		}
	}
	if !found {
		t.Error("expected 'Spec Mode' match in 'spec mode approval'")
	}
}
