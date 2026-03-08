---
status: Draft
approved_at: ""
approved_by: ""
---
# Spec 076-plan-decomposition-quality: Research-Informed Plan Decomposition Quality Gates

## Goal

Add deterministic decomposition quality analysis to `mindspec validate plan`, using metrics derived from Kim et al. (2025) "Towards a Science of Scaling Agent Systems" (arXiv:2512.08296). The validator computes scope redundancy, dependency graph shape, and bead count, then emits actionable warnings when a plan's decomposition structure correlates with known multi-agent performance degradation patterns.

## Background

### Research basis

Kim et al. evaluated 180 configurations of multi-agent systems across 5 architectures and 4 benchmarks. Their key findings relevant to plan decomposition:

1. **Sequential dependency chains degrade performance -39% to -70%** (PlanCraft benchmark). Tasks with strict sequential interdependence — where each step modifies state the next step depends on — universally degrade under multi-agent coordination. The coordination overhead fragments reasoning capacity without decomposition benefit.

2. **Optimal redundancy R≈0.41; R>0.50 hurts** (r=-0.136, p=0.004). Some overlap between work units is healthy for context continuity, but too much is wasteful. High redundancy means agents duplicate work rather than divide it.

3. **3-4 parallel agents is the sweet spot**. Turn scaling follows a power law with exponent 1.724 (super-linear). Per-agent reasoning capacity becomes prohibitively thin beyond 3-4 agents under fixed token budgets.

4. **Capability saturation at ~45% single-agent baseline**. When one agent can already handle a task, splitting it across multiple agents yields negative returns (β=-0.404, p<0.001). Over-decomposing trivial work creates coordination overhead with no benefit.

5. **Tool-heavy tasks suffer 6.3x efficiency penalty** under multi-agent fragmentation (β=-0.267, p<0.001). Token budgets fragment, leaving insufficient capacity per agent for complex tool reasoning.

### Applicability to mindspec

In mindspec, each bead is executed by a separate agent session (analogous to an independent agent in the paper). The plan's bead decomposition directly determines the multi-agent topology. Plans with deep serial chains, excessive bead counts, or high file overlap between beads match the paper's degradation patterns. These properties are statically computable from the plan markdown — no runtime data needed.

### Current state

`ValidatePlan()` in `internal/validate/plan.go` already parses bead sections (steps, verification, dependencies) and validates individual bead quality. It does not analyze the decomposition as a whole — no cross-bead analysis exists.

### What this spec adds

A `checkDecompositionQuality()` function that computes three metrics from the parsed `[]BeadSection` and emits warnings (not errors) when values fall outside empirically-grounded thresholds.

## Impacted Domains

- validate: New decomposition quality checks in `ValidatePlan()` after individual bead checks
- instruct: Plan-mode template updated with decomposition guidance referencing the research

## ADR Touchpoints

- [ADR-0016](../../adr/ADR-0016.md): Bead Creation Timing — this spec adds quality analysis to the plan validation that gates bead creation
- [ADR-0002](../../adr/ADR-0002.md): MindSpec + Beads integration — decomposition quality directly affects how beads are created from plans

## Requirements

1. `ParseBeadSections()` must capture raw step lines (`StepLines []string`) in addition to the existing `StepsCount`, so file paths can be extracted from step text
2. A new `ExtractPathRefs(text string) []string` function must extract file/package path references from arbitrary text using regex (matching patterns like `internal/foo/bar.go`, `cmd/mindspec/root.go`, `./internal/foo/...`, `go test ./pkg/...`)
3. `checkDecompositionQuality()` must compute and warn on:
   - **Scope redundancy (R_scope)**: `|paths referenced by >1 bead| / |total unique paths|`. Warn if R > 0.50 ("high bead overlap — consider merging beads that share most files") or R < 0.15 with >2 beads ("low overlap — beads may lack shared context")
   - **Dependency chain depth**: Longest path in the bead dependency DAG. Warn if depth > 3 ("deep serial chain — coordination overhead grows super-linearly")
   - **Parallelism ratio**: `beads with zero inbound deps / total beads`. Warn if < 0.25 ("most beads are serial — check for false dependencies")
   - **Bead count**: Warn if > 6 ("plan has N beads — consider whether decomposition is too fine-grained; 3-5 is optimal for agent coordination")
4. Dependency parsing must reuse the existing `bead\s+(\d+)` regex pattern from `internal/approve/plan.go` to build an adjacency list from `DependsOn` text
5. All decomposition warnings must be warnings (not errors) — they are advisory signals, not hard gates. Plans with legitimate reasons for deep chains or many beads should still pass validation.
6. The plan-mode instruct template (`internal/instruct/templates/plan.md`) must include decomposition guidance in the "Required Output" section, referencing the research-backed thresholds
7. `mindspec validate plan <id>` must include decomposition analysis in its output when warnings are present

## Scope

### In Scope
- `internal/validate/plan.go` — extend `ParseBeadSections()` to capture `StepLines`, add `ExtractPathRefs()`, add `checkDecompositionQuality()`
- `internal/validate/plan_test.go` — unit tests for path extraction, R_scope calculation, dependency graph analysis, and threshold warnings
- `internal/instruct/templates/plan.md` — add decomposition guidance section

### Out of Scope
- Runtime metrics (actual agent performance, token usage) — this spec uses static analysis only
- Changing the bead section format — the existing `## Bead N:` / `**Steps**` / `**Depends on**` format is sufficient
- Hard errors for any decomposition metric — all checks are warnings
- Changes to `internal/approve/plan.go` — bead creation logic is unchanged
- P_SA (single-agent baseline) estimation — would require historical data, not in scope

## Non-Goals

- Predicting actual agent performance — we provide directional warnings based on structural correlates, not performance predictions
- Enforcing a maximum bead count — teams may have legitimate reasons for larger plans
- Changing existing plans retroactively — checks skip already-approved plans (consistent with existing `isApproved` pattern)

## Acceptance Criteria

- [ ] `ParseBeadSections()` returns `StepLines []string` for each bead section
- [ ] `ExtractPathRefs()` correctly extracts Go file paths, package paths, and test paths from arbitrary text
- [ ] `mindspec validate plan` warns when R_scope > 0.50 (high overlap)
- [ ] `mindspec validate plan` warns when R_scope < 0.15 with >2 beads (low overlap)
- [ ] `mindspec validate plan` warns when dependency chain depth > 3
- [ ] `mindspec validate plan` warns when parallelism ratio < 0.25
- [ ] `mindspec validate plan` warns when bead count > 6
- [ ] All warnings include the computed metric value and the research-backed threshold
- [ ] Already-approved plans skip decomposition checks (consistent with existing pattern)
- [ ] Plan-mode instruct template includes decomposition guidance
- [ ] `go test ./internal/validate/...` passes with new test cases

## Validation Proofs

- `go test ./internal/validate/ -run TestExtractPathRefs`: path extraction from various text patterns
- `go test ./internal/validate/ -run TestDecompositionQuality`: all threshold warnings fire correctly
- `go test ./internal/validate/ -run TestDecompositionQuality_NoWarnings`: a well-structured plan produces no warnings
- `mindspec validate plan 074-self-contained-beads`: existing approved plan produces no decomposition errors (skipped)
- `./bin/mindspec instruct` in plan mode: includes decomposition guidance text

## Open Questions

- [x] Should decomposition checks be errors or warnings? **Resolved**: Warnings only. The thresholds are empirical correlates, not absolute rules. A 7-bead plan with legitimate need for each bead should not be blocked.
- [x] Should we compute R_scope from step text only, or also verification text? **Resolved**: Both. Steps describe what to change, verification describes what to test — both reference file paths that indicate scope overlap.
- [x] What path patterns to extract? **Resolved**: Go-specific patterns for now (`*.go`, `./internal/...`, `go test ./...`). The regex can be extended for other languages later without API changes.

## Approval

- **Status**: DRAFT
- **Approved By**: -
- **Approval Date**: -
- **Notes**: -
