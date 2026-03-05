---
status: Draft
approved_at: ""
approved_by: ""
---
# Spec 073-llm-test-coverage: Improve LLM test coverage and iteration

## Goal

Improve the reliability and coverage of the LLM behavioral test suite so that more scenarios pass consistently on Haiku, and remaining guidance gaps in the CLI/instruct templates are surfaced and fixed.

## Background

The LLM test harness (`internal/harness/`) runs 18 behavioral scenarios against real Claude Code sessions. As of 2026-03-04, the latest full-suite run shows **11 PASS, 7 FAIL**. All 7 failures are categorized as "pre-existing haiku behavior" — the agent doesn't follow the mindspec lifecycle reliably in certain scenarios.

Root causes fall into two categories:
1. **Guidance gaps** — instruct templates or CLI error messages don't steer the agent strongly enough (fix surface = `internal/instruct/templates/`, CLI output)
2. **Scenario design issues** — some scenarios are too complex for Haiku's turn budget, or assertions don't account for valid alternative paths

Per the fix-surface rule (MEMORY.md), fixes go into mindspec's own guidance — never into test prompts.

### Current failing scenarios (2026-03-04 setupWorktrees run):
- `SpecToIdle` — agent skips `mindspec complete` after `mindspec next`
- `ResumeAfterCrash` — agent uses raw git instead of `mindspec complete`
- `InterruptForBug` — agent commits directly to main (`skip_next` wrong action)
- `BugfixBranch` — agent commits to main, no branch/PR
- `MultipleActiveSpecs` — no `mindspec complete --spec`
- `StaleWorktree` — max turns exhausted
- `UnmergedBeadGuard` — setup failure (`bd create spec epic` exit 1)

### Current passing scenarios (11/18):
SingleBead, SpecInit, SpecApprove, PlanApprove, ImplApprove, SpecStatus, ApproveSpecFromWorktree, ApprovePlanFromWorktree, BlockedBeadTransition, MultiBeadDeps, CompleteFromSpecWorktree (intermittent)

## Impacted Domains

- `internal/harness/` — test scenarios, assertions, sandbox
- `internal/instruct/templates/` — agent guidance templates (primary fix surface)
- CLI error messages in `cmd/mindspec/` — secondary fix surface

## ADR Touchpoints

- [ADR-0023](../../adr/ADR-0023.md): Beads-based phase derivation affects scenario setup

## Requirements

1. Investigate each failing scenario to identify the root cause (guidance gap vs. scenario design)
2. Fix guidance gaps in instruct templates or CLI output to steer Haiku toward correct behavior
3. Adjust scenario assertions where valid alternative agent paths are rejected
4. Fix `UnmergedBeadGuard` setup failure (sandbox issue, not agent behavior)
5. `mindspec setup claude` should detect and remove stale git hooks (`.backup`, `.pre-mindspec` suffixed copies, removed hooks like `post-checkout`) — mirroring how it already detects and removes stale Claude Code hooks from `settings.json`
6. Fix `skip_next` analyzer false positives — `detectSkipNext()` currently fires in sessions that never enter implement phase (e.g. SpecInit, PlanApprove). The rule should bail out early if `mindspec next` never appears in the event stream AND no event has `Phase == "implement"`. Commits during spec/plan workflows are not violations. Also update `ApproveSpecFromWorktree` MaxTurns (10 is too low — agent runs out exploring help).
7. Strengthen assertions on simple scenarios: SpecApprove should assert mode transition to plan (verify `mindspec state show` reports plan mode — not focus files, which are retired). ApproveSpecFromWorktree needs `sandbox.CreateSpecEpic(specID)` in setup (missing — SpecApprove has it but this variant doesn't, likely causing `approve spec` to fail under ADR-0023 phase derivation) and richer assertions. ApprovePlanFromWorktree should check bead creation and branch state.
8. Clean up stale focus references in scenario.go — comments like "focus.activeWorktree", "Set focus to spec mode", and commit messages like `"setup: spec mode focus"` reference the retired focus file system. Update to reflect ADR-0023 beads-based phase derivation.
9. De-tautologize FromWorktree prompts — ApproveSpecFromWorktree and ApprovePlanFromWorktree prompts currently name the exact action ("Approve the spec/plan"). Replace with outcome-oriented prompts that let the agent discover the right command from `mindspec instruct` (e.g. "The spec/plan is finished. Advance the project to the next lifecycle phase.").
10. Track improvement via TESTING.md history tables (before/after per scenario)
11. No regressions in currently-passing scenarios

## Scope

### In Scope
- `internal/instruct/templates/*.md` — guidance improvements
- `internal/harness/scenario.go` — assertion adjustments for valid alternatives
- `internal/harness/analyzer.go` — fix `detectSkipNext()` false positives for non-implement sessions
- `internal/harness/analyzer_test.go` — deterministic tests for the fix
- `internal/harness/sandbox.go` — setup fixes (UnmergedBeadGuard)
- `cmd/mindspec/` CLI error messages — clearer recovery guidance
- `internal/harness/TESTING.md` — improvement history tracking
- `internal/hooks/install.go` — stale git hook cleanup (`.backup`, `.pre-mindspec`, removed hooks)

### Out of Scope
- Switching from Haiku to a more capable model
- Adding entirely new scenarios (focus on fixing existing ones)
- Refactoring the harness infrastructure

## Non-Goals

- 100% pass rate — some scenarios may remain flaky on Haiku due to model limitations
- Modifying test prompts to prescribe commands (violates fix-surface rule)

## Acceptance Criteria

- [ ] At least 3 currently-failing scenarios pass reliably (2+ consecutive runs)
- [ ] `UnmergedBeadGuard` setup failure is fixed
- [ ] No regressions in the 11 currently-passing scenarios
- [ ] TESTING.md updated with improvement history rows for each changed scenario
- [ ] All fixes are in guidance/CLI output, not test prompts
- [ ] `mindspec setup claude` removes stale git hooks (`.backup`, `.pre-mindspec`, dead `post-checkout`)
- [ ] `skip_next` no longer fires false positives in SpecInit, PlanApprove, or other non-implement sessions
- [ ] `ApproveSpecFromWorktree` MaxTurns increased to allow completion

## Validation Proofs

- `env -u CLAUDECODE go test ./internal/harness/ -v -run '^TestLLM_' -timeout 180m -count=1`: Full suite run with improved pass count
- `go test ./internal/harness/ -short -v`: All deterministic tests pass

## Open Questions

- [ ] Which failing scenarios are most impactable via guidance changes vs. inherent Haiku limitations?
- [ ] Should `StaleWorktree` MaxTurns be increased, or is the scenario too complex for Haiku?
- [ ] Is `BugfixBranch` a realistic test for Haiku (agent must discover branch policy from guidance alone)?

## Approval

- **Status**: DRAFT
- **Approved By**: -
- **Approval Date**: -
- **Notes**: Exploratory spec — scope may narrow during planning based on initial investigation
