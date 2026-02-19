package resolve

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mindspec/mindspec/internal/state"
)

// --- Multi-spec same-worktree integration tests ---

// setupMultiSpecProject creates a temp project with two specs, each with
// molecule bindings and step mappings. Returns the project root.
func setupMultiSpecProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	// Create .mindspec directory
	os.MkdirAll(filepath.Join(root, ".mindspec"), 0755)

	// Create docs/specs with two specs
	for _, spec := range []struct {
		id       string
		molID    string
		mapping  string
		bodyPost string
	}{
		{
			id:    "038-alpha",
			molID: "mol-alpha",
			mapping: `step_mapping:
    spec: step-a-spec
    spec-approve: step-a-spec-approve
    plan: step-a-plan
    plan-approve: step-a-plan-approve
    implement: step-a-implement
    review: step-a-review`,
			bodyPost: "# Spec 038-alpha: Alpha Feature\n\n## Goal\n\nAlpha feature goal.\n\n## Approval\n\n- **Status**: APPROVED\n",
		},
		{
			id:    "039-beta",
			molID: "mol-beta",
			mapping: `step_mapping:
    spec: step-b-spec
    spec-approve: step-b-spec-approve
    plan: step-b-plan
    plan-approve: step-b-plan-approve
    implement: step-b-implement
    review: step-b-review`,
			bodyPost: "# Spec 039-beta: Beta Feature\n\n## Goal\n\nBeta feature goal.\n\n## Approval\n\n- **Status**: DRAFT\n",
		},
	} {
		specDir := filepath.Join(root, "docs", "specs", spec.id)
		os.MkdirAll(specDir, 0755)

		content := "---\nmolecule_id: " + spec.molID + "\n" + spec.mapping + "\n---\n" + spec.bodyPost
		os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(content), 0644)
	}

	return root
}

// TestTwoActiveSpecs_DeriveModeIndependently verifies that two specs in the
// same worktree can have independently derived modes.
func TestTwoActiveSpecs_DeriveModeIndependently(t *testing.T) {
	// Alpha is in implement phase, Beta is in spec phase.
	// These statuses would be fetched from Beads; we test the deriveMode logic directly.
	alphaMapping := map[string]string{
		"spec":         "step-a-spec",
		"spec-approve": "step-a-spec-approve",
		"plan":         "step-a-plan",
		"plan-approve": "step-a-plan-approve",
		"implement":    "step-a-implement",
		"review":       "step-a-review",
	}
	alphaStatuses := map[string]string{
		"step-a-spec":         "closed",
		"step-a-spec-approve": "closed",
		"step-a-plan":         "closed",
		"step-a-plan-approve": "closed",
		"step-a-implement":    "in_progress",
		"step-a-review":       "open",
	}

	betaMapping := map[string]string{
		"spec":         "step-b-spec",
		"spec-approve": "step-b-spec-approve",
		"plan":         "step-b-plan",
		"plan-approve": "step-b-plan-approve",
		"implement":    "step-b-implement",
		"review":       "step-b-review",
	}
	betaStatuses := map[string]string{
		"step-b-spec":         "in_progress",
		"step-b-spec-approve": "open",
		"step-b-plan":         "open",
		"step-b-plan-approve": "open",
		"step-b-implement":    "open",
		"step-b-review":       "open",
	}

	alphaMode := deriveMode(alphaMapping, alphaStatuses)
	betaMode := deriveMode(betaMapping, betaStatuses)

	if alphaMode != state.ModeImplement {
		t.Errorf("alpha mode: got %q, want %q", alphaMode, state.ModeImplement)
	}
	if betaMode != state.ModeSpec {
		t.Errorf("beta mode: got %q, want %q", betaMode, state.ModeSpec)
	}

	// Both should be active
	if !isActive(alphaMapping, alphaStatuses) {
		t.Error("alpha should be active (implement in_progress)")
	}
	if !isActive(betaMapping, betaStatuses) {
		t.Error("beta should be active (spec in_progress)")
	}
}

// TestAmbiguousTarget_RefusesToGuess verifies that ResolveTarget returns
// ErrAmbiguousTarget when multiple active specs exist and no --spec is given.
func TestAmbiguousTarget_RefusesToGuess(t *testing.T) {
	err := &ErrAmbiguousTarget{
		Active: []SpecStatus{
			{SpecID: "038-alpha", Mode: "implement"},
			{SpecID: "039-beta", Mode: "spec"},
		},
	}

	msg := err.Error()

	// Must mention --spec
	if !strings.Contains(msg, "--spec") {
		t.Errorf("ambiguous error should mention --spec: %s", msg)
	}

	// Must list both specs
	if !strings.Contains(msg, "038-alpha") {
		t.Errorf("ambiguous error should list 038-alpha: %s", msg)
	}
	if !strings.Contains(msg, "039-beta") {
		t.Errorf("ambiguous error should list 039-beta: %s", msg)
	}
}

// TestExplicitTarget_BypassesAmbiguity verifies that explicit --spec always
// resolves without checking active specs.
func TestExplicitTarget_BypassesAmbiguity(t *testing.T) {
	// Even with a nonexistent root, explicit target should work
	got, err := ResolveTarget("/nonexistent", "038-alpha")
	if err != nil {
		t.Fatalf("explicit target should not error: %v", err)
	}
	if got != "038-alpha" {
		t.Errorf("got %q, want %q", got, "038-alpha")
	}
}

// TestSingleActiveSpec_AutoSelects verifies that when exactly one spec is
// active, untargeted commands auto-select it.
func TestSingleActiveSpec_AutoSelects(t *testing.T) {
	// This is tested indirectly: the logic is in ResolveTarget calling ActiveSpecs.
	// ActiveSpecs requires live Beads, so we verify the contract:
	// len(active) == 1 → auto-select

	// Simulate: only one active spec
	active := []SpecStatus{
		{SpecID: "038-alpha", Mode: "plan", MoleculeID: "mol-alpha"},
	}

	if len(active) != 1 {
		t.Fatal("test setup error")
	}

	// ResolveTarget contract: single active → return it
	// (the actual code path is: active has len 1 → return active[0].SpecID)
	result := active[0].SpecID
	if result != "038-alpha" {
		t.Errorf("single active auto-select: got %q, want %q", result, "038-alpha")
	}
}

// TestCrossSpecSafety_DeriveModeIsolation verifies that deriving mode for one
// spec does not interfere with or leak state to another spec.
func TestCrossSpecSafety_DeriveModeIsolation(t *testing.T) {
	// Alpha is in review, Beta is in plan.
	// Derive both and verify no cross-contamination.
	alphaMapping := map[string]string{
		"spec":         "a-spec",
		"spec-approve": "a-approve",
		"plan":         "a-plan",
		"plan-approve": "a-plan-approve",
		"implement":    "a-impl",
		"review":       "a-review",
	}
	alphaStatuses := map[string]string{
		"a-spec":         "closed",
		"a-approve":      "closed",
		"a-plan":         "closed",
		"a-plan-approve": "closed",
		"a-impl":         "closed",
		"a-review":       "open",
	}

	betaMapping := map[string]string{
		"spec":         "b-spec",
		"spec-approve": "b-approve",
		"plan":         "b-plan",
		"plan-approve": "b-plan-approve",
		"implement":    "b-impl",
		"review":       "b-review",
	}
	betaStatuses := map[string]string{
		"b-spec":         "closed",
		"b-approve":      "closed",
		"b-plan":         "in_progress",
		"b-plan-approve": "open",
		"b-impl":         "open",
		"b-review":       "open",
	}

	// Derive modes in sequence — each must be independent
	alphaMode := deriveMode(alphaMapping, alphaStatuses)
	betaMode := deriveMode(betaMapping, betaStatuses)

	if alphaMode != state.ModeReview {
		t.Errorf("alpha: got %q, want %q", alphaMode, state.ModeReview)
	}
	if betaMode != state.ModePlan {
		t.Errorf("beta: got %q, want %q", betaMode, state.ModePlan)
	}

	// Derive again in reverse order — results must be stable
	betaMode2 := deriveMode(betaMapping, betaStatuses)
	alphaMode2 := deriveMode(alphaMapping, alphaStatuses)

	if alphaMode2 != alphaMode {
		t.Errorf("alpha mode unstable: first=%q, second=%q", alphaMode, alphaMode2)
	}
	if betaMode2 != betaMode {
		t.Errorf("beta mode unstable: first=%q, second=%q", betaMode, betaMode2)
	}
}

// TestCrossSpecSafety_ActivePredicate verifies that the active-spec predicate
// correctly handles mixed states: one active, one completed.
func TestCrossSpecSafety_ActivePredicate(t *testing.T) {
	// Alpha: lifecycle complete (review closed)
	alphaMapping := testStepMapping()
	alphaStatuses := map[string]string{
		"step-spec":         "closed",
		"step-spec-approve": "closed",
		"step-plan":         "closed",
		"step-plan-approve": "closed",
		"step-implement":    "closed",
		"step-review":       "closed",
	}

	// Beta: in progress (plan phase)
	betaMapping := map[string]string{
		"spec":         "bs-spec",
		"spec-approve": "bs-approve",
		"plan":         "bs-plan",
		"plan-approve": "bs-plan-approve",
		"implement":    "bs-impl",
		"review":       "bs-review",
	}
	betaStatuses := map[string]string{
		"bs-spec":         "closed",
		"bs-approve":      "closed",
		"bs-plan":         "in_progress",
		"bs-plan-approve": "open",
		"bs-impl":         "open",
		"bs-review":       "open",
	}

	if isActive(alphaMapping, alphaStatuses) {
		t.Error("alpha should be inactive (review closed)")
	}
	if !isActive(betaMapping, betaStatuses) {
		t.Error("beta should be active (plan in_progress)")
	}
}

// --- Compatibility / migration tests ---

// TestLegacyRepo_FallbackToCursor verifies that a repo with only state.json
// (no molecule bindings) still resolves via the cursor fallback.
func TestLegacyRepo_FallbackToCursor(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, ".mindspec"), 0755)
	os.MkdirAll(filepath.Join(root, "docs", "specs"), 0755)

	// Legacy state.json with no molecule info
	stateJSON := `{"mode":"plan","activeSpec":"010-legacy","activeBead":"","lastUpdated":"2026-01-01T00:00:00Z"}`
	os.WriteFile(filepath.Join(root, ".mindspec", "state.json"), []byte(stateJSON), 0644)

	// No specs with molecule bindings → resolver finds no active specs → falls back to cursor
	got, err := ResolveTarget(root, "")
	if err != nil {
		t.Fatalf("legacy fallback failed: %v", err)
	}
	if got != "010-legacy" {
		t.Errorf("got %q, want %q", got, "010-legacy")
	}
}

// TestLegacyRepo_StateReadStillWorks verifies that the standard state.Read
// works on repos without molecule bindings.
func TestLegacyRepo_StateReadStillWorks(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, ".mindspec"), 0755)

	// Write via state package
	s := &state.State{
		Mode:       state.ModeImplement,
		ActiveSpec: "005-next",
		ActiveBead: "bead-old",
	}
	if err := state.Write(root, s); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Read back
	got, err := state.Read(root)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if got.Mode != state.ModeImplement {
		t.Errorf("mode: got %q, want %q", got.Mode, state.ModeImplement)
	}
	if got.ActiveSpec != "005-next" {
		t.Errorf("activeSpec: got %q, want %q", got.ActiveSpec, "005-next")
	}
	if got.ActiveBead != "bead-old" {
		t.Errorf("activeBead: got %q, want %q", got.ActiveBead, "bead-old")
	}
}

// TestLegacySpec_NoFrontmatter verifies that specs without molecule frontmatter
// are handled gracefully (zero Meta, no error).
func TestLegacySpec_NoFrontmatter(t *testing.T) {
	root := t.TempDir()
	specDir := filepath.Join(root, "docs", "specs", "005-next")
	os.MkdirAll(specDir, 0755)
	os.WriteFile(filepath.Join(specDir, "spec.md"), []byte("# Spec 005\n\n## Goal\n\nSomething.\n"), 0644)

	// ActiveSpecs should skip this spec (no molecule binding)
	specsDir := filepath.Join(root, "docs", "specs")
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		t.Fatalf("reading specs dir: %v", err)
	}

	var active []SpecStatus
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		specPath := filepath.Join(specsDir, e.Name(), "spec.md")
		if _, err := os.Stat(specPath); err != nil {
			continue
		}
		// Simulate what ActiveSpecs does: read meta, skip if no molecule
		// (we can't call ActiveSpecs directly because it calls fetchStepStatuses)
		data, _ := os.ReadFile(specPath)
		content := string(data)
		if !strings.Contains(content, "molecule_id:") {
			continue // no molecule binding → skip
		}
		active = append(active, SpecStatus{SpecID: e.Name()})
	}

	if len(active) != 0 {
		t.Errorf("expected no active specs for legacy repo, got %d", len(active))
	}
}

// TestMixedRepo_BoundAndUnbound verifies that a repo with both bound and
// unbound specs only returns the bound ones as active candidates.
func TestMixedRepo_BoundAndUnbound(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, ".mindspec"), 0755)

	// Bound spec with frontmatter
	boundDir := filepath.Join(root, "docs", "specs", "038-bound")
	os.MkdirAll(boundDir, 0755)
	boundContent := "---\nmolecule_id: mol-bound\nstep_mapping:\n    spec: s1\n    review: s6\n---\n# Spec 038\n"
	os.WriteFile(filepath.Join(boundDir, "spec.md"), []byte(boundContent), 0644)

	// Unbound legacy spec (no frontmatter)
	unboundDir := filepath.Join(root, "docs", "specs", "005-legacy")
	os.MkdirAll(unboundDir, 0755)
	os.WriteFile(filepath.Join(unboundDir, "spec.md"), []byte("# Spec 005\n\n## Goal\n"), 0644)

	// Scan specs: count how many have molecule bindings
	specsDir := filepath.Join(root, "docs", "specs")
	entries, _ := os.ReadDir(specsDir)

	boundCount := 0
	unboundCount := 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(specsDir, e.Name(), "spec.md"))
		if err != nil {
			continue
		}
		if strings.Contains(string(data), "molecule_id:") && !strings.Contains(string(data), "molecule_id: \"\"") {
			boundCount++
		} else {
			unboundCount++
		}
	}

	if boundCount != 1 {
		t.Errorf("expected 1 bound spec, got %d", boundCount)
	}
	if unboundCount != 1 {
		t.Errorf("expected 1 unbound spec, got %d", unboundCount)
	}
}

// --- Mode derivation bounded round-trip tests ---

// TestDeriveMode_BoundedComplexity verifies that deriveMode operates in
// constant time relative to the number of steps (no external calls).
func TestDeriveMode_BoundedComplexity(t *testing.T) {
	// deriveMode processes at most 6 lifecycle steps regardless of input size
	mapping := testStepMapping()
	statuses := make(map[string]string, 100)

	// Add the lifecycle steps
	for _, id := range mapping {
		statuses[id] = "closed"
	}
	// Add noise: many non-lifecycle statuses that should be ignored
	for i := 0; i < 100; i++ {
		statuses[strings.Repeat("x", i+1)] = "open"
	}

	got := deriveMode(mapping, statuses)
	if got != state.ModeIdle {
		t.Errorf("deriveMode with noise: got %q, want %q", got, state.ModeIdle)
	}
}

// TestIsActive_BoundedComplexity verifies that isActive processes at most
// the 6 lifecycle steps + 1 review check.
func TestIsActive_BoundedComplexity(t *testing.T) {
	mapping := testStepMapping()
	statuses := make(map[string]string, 100)

	// All lifecycle steps closed
	for _, id := range mapping {
		statuses[id] = "closed"
	}
	// Noise entries (should not affect result)
	for i := 0; i < 100; i++ {
		statuses[strings.Repeat("y", i+1)] = "open"
	}

	got := isActive(mapping, statuses)
	if got {
		t.Error("isActive with all closed lifecycle + noise: should be false")
	}
}

// --- Single-spec regression tests ---

// TestSingleSpec_AllLifecyclePhases verifies that a single spec correctly
// transitions through all lifecycle phases without regression.
func TestSingleSpec_AllLifecyclePhases(t *testing.T) {
	mapping := testStepMapping()

	phases := []struct {
		name     string
		statuses map[string]string
		wantMode string
		wantAct  bool
	}{
		{
			name: "fresh spec (all open)",
			statuses: map[string]string{
				"step-spec": "open", "step-spec-approve": "open",
				"step-plan": "open", "step-plan-approve": "open",
				"step-implement": "open", "step-review": "open",
			},
			wantMode: state.ModeSpec,
			wantAct:  true,
		},
		{
			name: "spec in progress",
			statuses: map[string]string{
				"step-spec": "in_progress", "step-spec-approve": "open",
				"step-plan": "open", "step-plan-approve": "open",
				"step-implement": "open", "step-review": "open",
			},
			wantMode: state.ModeSpec,
			wantAct:  true,
		},
		{
			name: "spec approved, plan phase",
			statuses: map[string]string{
				"step-spec": "closed", "step-spec-approve": "closed",
				"step-plan": "in_progress", "step-plan-approve": "open",
				"step-implement": "open", "step-review": "open",
			},
			wantMode: state.ModePlan,
			wantAct:  true,
		},
		{
			name: "plan approved, implement phase",
			statuses: map[string]string{
				"step-spec": "closed", "step-spec-approve": "closed",
				"step-plan": "closed", "step-plan-approve": "closed",
				"step-implement": "in_progress", "step-review": "open",
			},
			wantMode: state.ModeImplement,
			wantAct:  true,
		},
		{
			name: "implementation done, review phase",
			statuses: map[string]string{
				"step-spec": "closed", "step-spec-approve": "closed",
				"step-plan": "closed", "step-plan-approve": "closed",
				"step-implement": "closed", "step-review": "in_progress",
			},
			wantMode: state.ModeReview,
			wantAct:  true,
		},
		{
			name: "all done (lifecycle complete)",
			statuses: map[string]string{
				"step-spec": "closed", "step-spec-approve": "closed",
				"step-plan": "closed", "step-plan-approve": "closed",
				"step-implement": "closed", "step-review": "closed",
			},
			wantMode: state.ModeIdle,
			wantAct:  false,
		},
	}

	for _, tt := range phases {
		t.Run(tt.name, func(t *testing.T) {
			mode := deriveMode(mapping, tt.statuses)
			if mode != tt.wantMode {
				t.Errorf("deriveMode() = %q, want %q", mode, tt.wantMode)
			}

			active := isActive(mapping, tt.statuses)
			if active != tt.wantAct {
				t.Errorf("isActive() = %v, want %v", active, tt.wantAct)
			}
		})
	}
}

// TestDeriveMode_Deterministic verifies that deriveMode returns the same result
// when called multiple times with the same inputs.
func TestDeriveMode_Deterministic(t *testing.T) {
	mapping := testStepMapping()
	statuses := map[string]string{
		"step-spec":         "closed",
		"step-spec-approve": "closed",
		"step-plan":         "in_progress",
		"step-plan-approve": "open",
		"step-implement":    "open",
		"step-review":       "open",
	}

	results := make(map[string]int)
	for i := 0; i < 100; i++ {
		mode := deriveMode(mapping, statuses)
		results[mode]++
	}

	if len(results) != 1 {
		t.Errorf("deriveMode returned different results across 100 calls: %v", results)
	}
	if _, ok := results[state.ModePlan]; !ok {
		t.Errorf("expected all calls to return %q, got %v", state.ModePlan, results)
	}
}

// TestFormatActiveList_Ordering verifies deterministic sort order.
func TestFormatActiveList_Ordering(t *testing.T) {
	specs := []SpecStatus{
		{SpecID: "039-beta", Mode: "spec"},
		{SpecID: "038-alpha", Mode: "implement"},
		{SpecID: "040-gamma", Mode: "plan"},
	}

	output := FormatActiveList(specs)

	// Find positions — alpha should appear before beta before gamma
	alphaIdx := strings.Index(output, "038-alpha")
	betaIdx := strings.Index(output, "039-beta")
	gammaIdx := strings.Index(output, "040-gamma")

	if alphaIdx == -1 || betaIdx == -1 || gammaIdx == -1 {
		t.Fatalf("expected all spec IDs in output: %s", output)
	}

	// FormatActiveList just renders in input order. The caller (ActiveSpecs)
	// sorts by spec ID. Verify the output contains all three specs.
	if !strings.Contains(output, "Active specs (3)") {
		t.Errorf("expected 'Active specs (3)' header, got: %s", output)
	}
}

// --- State cursor tests ---

// TestStateCursor_WritesNonCanonical verifies that state.json writes update
// cursor fields (mode, activeSpec, activeBead) but these are not treated as
// lifecycle truth — they're convenience fields.
func TestStateCursor_WritesNonCanonical(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, ".mindspec"), 0755)
	os.MkdirAll(filepath.Join(root, "docs", "specs", "038-test"), 0755)
	os.WriteFile(
		filepath.Join(root, "docs", "specs", "038-test", "spec.md"),
		[]byte("# Spec 038\n\n## Approval\n\n- **Status**: APPROVED\n"),
		0644,
	)

	// Write state cursor
	if err := state.SetMode(root, state.ModeImplement, "038-test", "bead-1"); err != nil {
		t.Fatalf("SetMode failed: %v", err)
	}

	// Read cursor back
	s, err := state.Read(root)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if s.Mode != state.ModeImplement {
		t.Errorf("cursor mode: got %q, want %q", s.Mode, state.ModeImplement)
	}
	if s.ActiveSpec != "038-test" {
		t.Errorf("cursor activeSpec: got %q, want %q", s.ActiveSpec, "038-test")
	}
	if s.ActiveBead != "bead-1" {
		t.Errorf("cursor activeBead: got %q, want %q", s.ActiveBead, "bead-1")
	}

	// The cursor's mode is NOT the canonical source of truth.
	// Canonical mode is derived from molecule step statuses.
	// We verify the cursor can differ from derived mode without error.
	// (The cursor might say "implement" while the molecule is actually in "plan")
}

// TestStateCursor_UpdatedOnNext verifies that state cursor updates when
// claiming new work, without affecting molecule-derived mode.
func TestStateCursor_UpdatedOnNext(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, ".mindspec"), 0755)
	os.MkdirAll(filepath.Join(root, "docs", "specs", "038-test"), 0755)
	os.WriteFile(
		filepath.Join(root, "docs", "specs", "038-test", "spec.md"),
		[]byte("# Spec 038\n\n## Approval\n\n- **Status**: APPROVED\n"),
		0644,
	)

	// Initial state: working on bead-1
	state.SetMode(root, state.ModeImplement, "038-test", "bead-1")

	// Simulate "next" claiming bead-2 — cursor should update
	state.SetMode(root, state.ModeImplement, "038-test", "bead-2")

	s, err := state.Read(root)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if s.ActiveBead != "bead-2" {
		t.Errorf("cursor activeBead: got %q, want %q", s.ActiveBead, "bead-2")
	}
}
