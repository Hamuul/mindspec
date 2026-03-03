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
version: 3
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
| `mindspec complete "msg"` auto-commits dirty worktree | Bead 1 (done) |
| `mindspec complete` with dirty tree fails with hint | Bead 1 (done) |
| `mindspec spec create` creates branch + worktree + template | Bead 2 (done) |
| `spec-init` still works as hidden alias | Bead 2 (done) |
| Phase-first commands (`spec approve`, `plan approve`, `impl approve`) | Bead 2 (done) |
| Old command forms remain as hidden aliases | Bead 2 (done) |
| All templates contain lifecycle map with `>>>` marker | Bead 3 (done) |
| No template references raw git as normal workflow | Bead 3 (done) |
| Explore mode fully removed (command, package, template, constant) | Bead 3 (done) |
| Harness scenarios updated for new commands | Bead 4 (done) |
| WORKFLOW-STATE-MACHINE.md updated | Bead 4 (done) |
| Spec updated to reflect actual implementation | Bead 4 (done) |
| `make test` passes | Bead 4 (done) |
| Stale CLI messages updated (spec-init refs, raw git in recovery hints) | Bead 5 |
| `mindspec next` enforces spec worktree scoping | Bead 6 |
| `mindspec complete` enforces bead worktree scoping | Bead 6 |
| `TestLLM_SingleBead` passes without raw git | Bead 7 |
| `TestLLM_SpecToIdle` passes full lifecycle | Bead 7 |

## Bead 1: Auto-commit in mindspec complete (DONE)

Add optional commit message parameter to `complete.Run()`. Before the clean-tree check, if a commit message is provided, call `gitops.CommitAll()` to stage and commit all changes. Update the dirty-tree error hint.

**Steps**
1. Add `commitMsg` string parameter to `complete.Run()` signature
2. Before `checkCleanWorktree()`, add auto-commit logic: resolve commit path (worktree or root), call `gitops.CommitAll(path, fmt.Sprintf("impl(%s): %s", beadID, commitMsg))` when commitMsg is non-empty
3. Update the dirty-tree error message hint to suggest `mindspec complete "describe what you did"`
4. In `cmd/mindspec/complete.go`, change `Args` to accept optional positional commit message, pass to `complete.Run()`
5. Add `commitAllFn` function variable for testability, wire to `gitops.CommitAll`

**Verification**
- [x] `go test ./internal/complete/ -v` passes
- [x] `make build && make test` passes

**Depends on**
None

## Bead 2: CLI namespace reorganization (DONE)

Create `spec`, `plan`, `impl` parent commands with subcommands. Keep backward-compat aliases hidden.

**Steps**
1. Create `cmd/mindspec/spec.go` with `specCmd` parent, `specCreateCmd`, `specApproveCmd`
2. Create `cmd/mindspec/plan_cmd.go` with `planCmd` parent, `planApproveCmd`
3. Create `cmd/mindspec/impl.go` with `implCmd` parent, `implApproveCmd`
4. Rewrite `cmd/mindspec/approve.go` to use shared RunE functions, mark as `Hidden: true`
5. Mark `specInitCmd` as `Hidden: true`
6. Register new commands in `root.go`
7. Update `CLAUDE.md` managed section

**Verification**
- [x] `make build && make test` passes
- [x] `mindspec spec create --help` shows usage

**Depends on**
None

## Bead 3: Instruct template updates and explore removal (DONE)

Update all instruct templates with lifecycle map. Soften raw git references to convention. Remove explore entirely.

**Steps**
1. Add lifecycle map with phase-specific `>>>` marker and git convention to all 5 templates
2. Delete explore template, command, package, state constant
3. Remove all explore references from hooks, instruct, bootstrap, setup, tests
4. Update idle template: remove `state set`, use `mindspec next` for resume

**Verification**
- [x] `make build && make test` passes
- [x] No explore references in any `.go` source file

**Depends on**
Bead 1, Bead 2

## Bead 4: Harness scenario updates, docs, and spec (DONE)

Update LLM test scenarios, WORKFLOW-STATE-MACHINE.md, and spec for new commands and explore removal.

**Steps**
1. Remove `ScenarioAbandonSpec` and `TestLLM_AbandonSpec`
2. Update remaining scenario assertions to accept both old and new command forms
3. Rewrite WORKFLOW-STATE-MACHINE.md: remove explore, add known gaps, soften git policy
4. Update 058 spec.md to reflect actual implementation

**Verification**
- [x] `make build && make test` passes

**Depends on**
Bead 1, Bead 2, Bead 3

## Bead 5: Stale CLI message cleanup

Several CLI error/help messages still reference `spec-init` and raw git commands. Update them to use the new command names and mindspec-native recovery paths.

**Steps**
1. `cmd/mindspec/next.go:75-76` — dirty tree recovery: replace `git add -A && git commit -m "wip"` / `git checkout -- .` with `mindspec complete "wip"` to auto-commit, or `git restore .` for discard (git restore is a repair action, acceptable)
2. `cmd/mindspec/next.go:130` — no-work message: replace `mindspec spec-init` with `mindspec spec create`
3. `cmd/mindspec/bead.go:21,25` — deprecated bead command: replace `mindspec spec-init` with `mindspec spec create`
4. `cmd/mindspec/instruct.go:157` — no-state fallback: replace `mindspec spec-init` with `mindspec spec create`
5. `cmd/mindspec/setup.go:31,75,115` — help text: replace `explore, spec-init` references with current command names

**Verification**
- [ ] `grep -r 'spec-init' cmd/mindspec/` only shows the hidden alias definition in `spec_init.go`
- [ ] `grep -r 'git add\|git commit\|git checkout' cmd/mindspec/` returns no results (except git restore for discard)
- [ ] `make build && make test` passes

**Depends on**
Bead 2

## Bead 6: Enforce worktree scoping for next and complete

`mindspec next` should only run from a spec worktree. `mindspec complete` should only run from a bead worktree. Add guards that detect the current worktree context and fail with a helpful message if run from the wrong location.

**Steps**
1. Add a `DetectWorktreeContext(cwd string) (kind string, specID string, beadID string)` helper that identifies whether the CWD is: main repo, spec worktree, or bead worktree (by path pattern matching against `.worktrees/worktree-spec-*` and `.worktrees/worktree-*`)
2. In `cmd/mindspec/next.go`, before step 1 (clean tree check), verify CWD is a spec worktree. If on main, error: "mindspec next must run from a spec worktree. Use `mindspec spec create <slug>` first." If in a bead worktree, error: "you're already in a bead worktree — run `mindspec complete "msg"` when done."
3. In `cmd/mindspec/complete.go` (or `internal/complete/complete.go`), verify CWD is a bead worktree. If on main or spec worktree, error with guidance.
4. Add `--allow-main` or similar escape hatch for recovery scenarios
5. Update tests to mock CWD where needed

**Verification**
- [ ] `mindspec next` from main repo fails with helpful error
- [ ] `mindspec next` from spec worktree succeeds
- [ ] `mindspec complete` from bead worktree succeeds
- [ ] `mindspec complete` from spec worktree fails with helpful error
- [ ] `make build && make test` passes

**Depends on**
Bead 5

## Bead 7: LLM harness integration verification

Run LLM harness tests to validate the full lifecycle works end-to-end with only mindspec commands.

**Steps**
1. `make build` to ensure binary is current
2. Run `env -u CLAUDECODE go test ./internal/harness/ -v -run TestLLM_SingleBead -timeout 10m -count=1`
3. If failures, diagnose and fix (per ADR-0021: fixes go into guidance, not test prompts)
4. Run `env -u CLAUDECODE go test ./internal/harness/ -v -run TestLLM_SpecToIdle -timeout 15m -count=1`
5. If failures, iterate
6. Run full suite: `env -u CLAUDECODE go test ./internal/harness/ -v -run '^TestLLM_' -timeout 180m -count=1`

**Verification**
- [ ] `TestLLM_SingleBead` passes
- [ ] `TestLLM_SpecToIdle` passes
- [ ] No agent runs raw git commands in successful test runs

**Depends on**
Bead 5, Bead 6
