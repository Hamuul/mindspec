---
approved_at: ""
approved_by: ""
bead_ids: []
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

## Bead 1: Auto-commit in `mindspec complete`

Add optional commit message parameter to `complete.Run()`. Before the clean-tree check, if a commit message is provided, call `gitops.CommitAll()` to stage and commit all changes. Update the dirty-tree error hint to suggest `mindspec complete "describe what you did"`.

**Files:**
- `internal/complete/complete.go` — add `commitMsg` param to `Run()`, insert auto-commit before clean-tree check
- `cmd/mindspec/complete.go` — wire positional arg for commit message

**Verification:** `go test ./internal/complete/ -v`

## Bead 2: CLI namespace reorganization

Create `spec` parent command with `create` and `approve` subcommands. Simplify explore to not change state. Keep backward-compat aliases hidden.

**Files:**
- `cmd/mindspec/spec.go` (NEW) — `specCmd` parent with `create` + `approve` subcommands
- `cmd/mindspec/plan.go` (NEW) — `planCmd` parent with `approve` subcommand
- `cmd/mindspec/impl.go` (NEW) — `implCmd` parent with `approve` subcommand
- `cmd/mindspec/spec_init.go` — mark hidden or remove (replaced by `spec create`)
- `cmd/mindspec/approve.go` — mark subcommands as `Hidden: true`
- `cmd/mindspec/explore.go` — `Enter()` no longer changes state, `promote` delegates to `spec create`
- `internal/explore/explore.go` — remove state writes from `Enter()`, `Dismiss()`, `Promote()`
- `cmd/mindspec/root.go` — register new command tree
- `CLAUDE.md` — update managed section

**Verification:** `make build && make test`

## Bead 3: Instruct template updates

Update all 6 instruct templates: add lifecycle map with phase-specific `>>>` marker, remove all raw git references, update command names.

**Files:**
- `internal/instruct/templates/idle.md` — merge explore guidance, add lifecycle map, remove session close git
- `internal/instruct/templates/explore.md` — stub redirecting to idle or remove
- `internal/instruct/templates/spec.md` — new command names, lifecycle map, no raw git
- `internal/instruct/templates/plan.md` — new command names, lifecycle map, no raw git
- `internal/instruct/templates/implement.md` — `mindspec complete "msg"`, lifecycle map, no raw git
- `internal/instruct/templates/review.md` — new command names, lifecycle map, no raw git

**Verification:** `make build && mindspec instruct` in each mode

## Bead 4: Harness scenario updates + integration verification

Update LLM test scenarios for new command names and assertions. Run full test suite.

**Files:**
- `internal/harness/scenario.go` — update prompts and assertions for new commands

**Verification:**
- `make build && make test`
- `go test ./internal/harness/ -short -v`
- `env -u CLAUDECODE go test ./internal/harness/ -v -run TestLLM_SingleBead -timeout 10m -count=1`
