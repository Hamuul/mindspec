package adr

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestShow(t *testing.T) {
	root := setupTestADRs(t)

	a, err := Show(root, "ADR-0001")
	if err != nil {
		t.Fatalf("Show: %v", err)
	}
	if a.Title != "Test Decision" {
		t.Errorf("Title = %q, want %q", a.Title, "Test Decision")
	}
	if a.Status != "Accepted" {
		t.Errorf("Status = %q, want Accepted", a.Status)
	}
}

func TestShow_NotFound(t *testing.T) {
	root := setupTestADRs(t)

	_, err := Show(root, "ADR-9999")
	if err == nil {
		t.Error("expected error for nonexistent ADR")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, want 'not found'", err.Error())
	}
}

func TestExtractDecision(t *testing.T) {
	content := `# ADR-0001: Test

## Context
Some context.

## Decision
We will use Redis for caching.

This improves performance.

## Consequences
Something.
`
	decision := ExtractDecision(content)
	if !strings.Contains(decision, "Redis for caching") {
		t.Errorf("decision = %q, expected Redis content", decision)
	}
	if strings.Contains(decision, "Consequences") {
		t.Error("decision should not include Consequences section")
	}
}

func TestExtractDecision_NoSection(t *testing.T) {
	content := "# ADR-0001: Test\n\n## Context\nJust context.\n"
	decision := ExtractDecision(content)
	if decision != "" {
		t.Errorf("expected empty decision, got %q", decision)
	}
}

func TestFormatSummary(t *testing.T) {
	root := setupTestADRs(t)
	a, _ := Show(root, "ADR-0001")

	summary := FormatSummary(a)
	if !strings.Contains(summary, "ADR-0001") {
		t.Error("expected ID in summary")
	}
	if !strings.Contains(summary, "Test Decision") {
		t.Error("expected title in summary")
	}
	if !strings.Contains(summary, "Accepted") {
		t.Error("expected status in summary")
	}
	if !strings.Contains(summary, "Decision:") {
		t.Error("expected Decision section in summary")
	}
}

func TestFormatJSON(t *testing.T) {
	root := setupTestADRs(t)
	a, _ := Show(root, "ADR-0001")

	jsonStr, err := FormatJSON(a)
	if err != nil {
		t.Fatalf("FormatJSON: %v", err)
	}

	// Verify it's valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if parsed["id"] != "ADR-0001" {
		t.Errorf("id = %v, want ADR-0001", parsed["id"])
	}
	if parsed["status"] != "Accepted" {
		t.Errorf("status = %v, want Accepted", parsed["status"])
	}
	if parsed["title"] != "Test Decision" {
		t.Errorf("title = %v, want Test Decision", parsed["title"])
	}
}

func TestFormatJSON_SupersededADR(t *testing.T) {
	root := setupTestADRs(t)
	a, _ := Show(root, "ADR-0003")

	jsonStr, err := FormatJSON(a)
	if err != nil {
		t.Fatalf("FormatJSON: %v", err)
	}

	var parsed map[string]interface{}
	json.Unmarshal([]byte(jsonStr), &parsed)

	if parsed["superseded_by"] != "ADR-0005" {
		t.Errorf("superseded_by = %v, want ADR-0005", parsed["superseded_by"])
	}
}
