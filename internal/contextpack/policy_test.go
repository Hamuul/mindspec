package contextpack

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParsePolicies(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "policies.yml")

	content := `# Header comment

policies:
  - id: test-policy-1
    description: "Test policy 1"
    severity: error
    scope: "src/**/*"
    mode: spec
    reference: "docs/test.md"

  - id: test-policy-2
    description: "Test policy 2"
    severity: warning

  - id: test-policy-3
    description: "Implementation only"
    severity: error
    mode: implementation
`
	os.WriteFile(path, []byte(content), 0o644)

	policies, err := ParsePolicies(path)
	if err != nil {
		t.Fatalf("ParsePolicies: %v", err)
	}

	if len(policies) != 3 {
		t.Fatalf("got %d policies, want 3", len(policies))
	}

	if policies[0].ID != "test-policy-1" {
		t.Errorf("policies[0].ID = %q", policies[0].ID)
	}
	if policies[0].Mode != "spec" {
		t.Errorf("policies[0].Mode = %q", policies[0].Mode)
	}
}

func TestFilterPolicies(t *testing.T) {
	policies := []Policy{
		{ID: "p1", Mode: "spec"},
		{ID: "p2", Mode: ""},
		{ID: "p3", Mode: "implementation"},
		{ID: "p4", Mode: "plan"},
	}

	filtered := FilterPolicies(policies, "spec")
	if len(filtered) != 2 {
		t.Fatalf("got %d filtered policies, want 2", len(filtered))
	}

	// Should contain p1 (mode=spec) and p2 (mode="" = all modes)
	ids := map[string]bool{}
	for _, p := range filtered {
		ids[p.ID] = true
	}
	if !ids["p1"] || !ids["p2"] {
		t.Errorf("expected p1 and p2, got %v", filtered)
	}
}

func TestFilterPolicies_AllModes(t *testing.T) {
	policies := []Policy{
		{ID: "p1", Mode: ""},
	}
	for _, mode := range []string{"spec", "plan", "implementation"} {
		filtered := FilterPolicies(policies, mode)
		if len(filtered) != 1 {
			t.Errorf("mode %q: got %d, want 1", mode, len(filtered))
		}
	}
}
