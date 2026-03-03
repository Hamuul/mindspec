---
approved_at: "2026-03-03T09:21:20Z"
approved_by: user
bead_ids:
    - mindspec-d5lm.1
    - mindspec-d5lm.2
    - mindspec-d5lm.3
    - mindspec-d5lm.4
last_updated: 2026-03-03T00:00:00Z
spec_id: 058-zero-git-lifecycle
status: Approved
version: 1
---

# Plan: Spec 058 — Zero Raw Git Lifecycle

## ADR Fitness

- **ADR-0006 (Worktree-first spec-init)**: Still sound. `spec create` reuses `specinit.Run()` unchanged — same worktree-first flow.
- **ADR-0022 (Worktree-aware resolution)**: Still sound. `complete`'s auto-commit uses the same worktree-aware path resolution.

## Testing Strategy

- Unit tests for `complete.Run()` with commitMsg parameter (auto-commit path)
- `make build && make test` — all existing tests pass
- `go test ./internal/harness/ -short -v` — deterministic harness tests
- LLM harness: `TestLLM_SingleBead` validates agent uses `mindspec complete "msg"` without raw git
- LLM harness: `TestLLM_SpecToIdle` validates full lifecycle with only mindspec commands

## Provenance

| Acceptance Criterion | Bead |
|---------------------|------|
| `mindspec complete "msg"` auto-commits dirty worktree | Bead 1 |
| `mindspec complete` with dirty tree fails with hint | Bead 1 |
| `mindspec spec create` creates branch + worktree + template | Bead 2 |
| `spec-init` still works as hidden alias | Bead 2 |
| `mindspec explore` does NOT change mode | Bead 2 |
| `mindspec explore promote` delegates to `spec create` | Bead 2 |
| All templates contain lifecycle map with `>>>` marker | Bead 3 |
| No template contains raw git command instructions | Bead 3 |
| Harness scenarios updated for new commands | Bead 4 |
| `make test` passes | Bead 4 |
| `TestLLM_SingleBead` passes without raw git | Bead 4 |
| `TestLLM_SpecToIdle` passes full lifecycle | Bead 4 |

## Bead 1: Auto-commit in mindspec complete

Add optional commit message parameter to `complete.Run()`. Before the clean-tree check, if a commit message is provided, call `gitops.CommitAll()` to stage and commit all changes. Update the dirty-tree error hint.

**Steps**
1. Add `commitMsg` string parameter to `complete.Run()` signature
2. Before `checkCleanWorktree()`, add auto-commit logic: resolve commit path (worktree or root), call `gitops.CommitAll(path, fmt.Sprintf("impl(%s): %s", beadID, commitMsg))` when commitMsg is non-empty
3. Update the dirty-tree error message hint to suggest `mindspec complete "describe what you did"`
4. In `cmd/mindspec/complete.go`, change `Args` to accept optional positional commit message, pass to `complete.Run()`
5. Add `commitAllFn` function variable for testability, wire to `gitops.CommitAll`

**Verification**
- [ ] `go test ./internal/complete/ -v` passes
- [ ] `make build && make test` passes

**Depends on**
None

## Bead 2: CLI namespace reorganization

Create `spec`, `plan`, `impl` parent commands with subcommands. Simplify explore. Keep backward-compat aliases hidden.

**Steps**
1. Create `cmd/mindspec/spec.go` with `specCmd` parent, `specCreateCmd` (wraps `specinit.Run()`), `specApproveCmd` (wraps existing approve-spec logic)
2. Create `cmd/mindspec/plan_cmd.go` with `planCmd` parent, `planApproveCmd` (wraps existing approve-plan logic)
3. Create `cmd/mindspec/impl.go` with `implCmd` parent, `implApproveCmd` (wraps existing approve-impl logic)
4. In `cmd/mindspec/approve.go`, mark all subcommands as `Hidden: true`; in `spec_init.go`, mark `specInitCmd` as `Hidden: true`
5. In `cmd/mindspec/root.go`, register `specCmd`, `planCmd`, `implCmd` as top-level commands
6. Simplify `internal/explore/explore.go`: `Enter()` removes state write (returns nil), `Dismiss()` removes mode check (returns nil), `Promote()` removes mode check
7. Update `CLAUDE.md` managed section with new command surface

**Verification**
- [ ] `make build && make test` passes
- [ ] `mindspec spec create --help` shows usage
- [ ] `mindspec spec-init --help` still works (hidden alias)

**Depends on**
None

## Bead 3: Instruct template updates

Update all 6 instruct templates with lifecycle map, remove raw git references, update command names.

**Steps**
1. Create lifecycle map block (common to all templates, with phase-specific `>>>` marker)
2. Update `idle.md`: merge explore guidance into subsection, add lifecycle map, remove session close git instructions
3. Update `explore.md`: simplify to guidance-only stub redirecting to idle
4. Update `spec.md`, `plan.md`: replace old command names, add lifecycle map, remove session close git
5. Update `implement.md`: replace commit convention with `mindspec complete "msg"`, add lifecycle map + git prohibition
6. Update `review.md`: replace old command names, add lifecycle map, remove session close git

**Verification**
- [ ] `make build` passes
- [ ] `make test` passes (template rendering tests)
- [ ] No raw git commands in any template (grep verification)

**Depends on**
Bead 1, Bead 2

## Bead 4: Harness scenario updates and integration verification

Update LLM test scenario prompts and assertions for new command names. Run full test suite.

**Steps**
1. Update `ScenarioSpecToIdle` prompt and assertions: `spec-init` → `spec create`, `approve spec` → `spec approve` (accept both forms)
2. Update `ScenarioSingleBead` assertions for `mindspec complete "msg"` pattern
3. Update `ScenarioAbandonSpec` for explore-without-mode-change behavior
4. Update remaining scenarios (`ScenarioSpecInit`, `ScenarioSpecApprove`, etc.) for new command forms
5. Run `make build && make test` and `go test ./internal/harness/ -short -v`

**Verification**
- [ ] `make build && make test` passes
- [ ] `go test ./internal/harness/ -short -v` passes
- [ ] `env -u CLAUDECODE go test ./internal/harness/ -v -run TestLLM_SingleBead -timeout 10m -count=1` passes
