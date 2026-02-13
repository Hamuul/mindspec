# Spec 008: Worktree Lifecycle Management

## Goal

Make worktree management implicit in MindSpec's happy-path workflow. Worktree CRUD is delegated to Beads (`bd worktree`). MindSpec orchestrates when worktrees are created and removed by integrating them into `mindspec next` (entry) and a new `mindspec complete` (exit). No standalone `mindspec worktree` commands — users use `bd worktree list/info/remove` directly for inspection and recovery.

## Background

Spec 007 introduced `mindspec bead worktree <bead-id> [--create]` as a thin wrapper around `git worktree add`. It was explicitly scoped to create/show only, deferring lifecycle management to this spec (007 spec.md, Out of Scope line 79).

The current gaps:

1. **Worktrees are not part of the happy path** — agents must manually create worktrees after claiming a bead and manually clean them after completion.
2. **No completion command** — bead completion is a multi-step manual checklist (verify, doc-sync, bd close, state advance, commit). There's no single command to orchestrate the close-out.
3. **MindSpec reimplements worktree CRUD** — `internal/bead/worktree.go` parses `git worktree list --porcelain` and calls `git worktree add` directly, bypassing Beads' worktree infrastructure. Beads provides `bd worktree create/list/remove/info` with redirect setup, safety checks, and daemon-awareness.
4. **`mindspec next` doesn't create worktrees** — it claims a bead and sets state, but the agent must separately create the worktree.

### Design principle

Beads owns worktree CRUD (`bd worktree create/list/remove/info`). MindSpec owns workflow orchestration (when to create, when to remove, naming conventions, state management). This aligns with ADR-0002 (Beads as passive substrate) and ADR-0003 (MindSpec owns orchestration).

### Existing code to replace

- `internal/bead/worktree.go` — `ParseWorktreeList()`, `FindWorktree()`, `CreateWorktree()`, `checkCleanTree()`. These reimplement functionality that `bd worktree` provides. Should be replaced with `bd worktree` calls.
- `internal/instruct/worktree.go` — `CheckWorktree()`. Can be simplified to use `bd worktree info`.
- `cmd/mindspec/bead.go` — `beadWorktreeCmd`. Deprecated in favor of `bd worktree` directly.

### Parallel workstreams

ADR-0007 (Proposed) explores per-worktree state to enable parallel spec/plan workstreams. This spec is designed so that its workflow integration extends naturally to spec/plan phases once ADR-0007 is accepted, but the v1 scope is implementation beads only.

## Impacted Domains

- workflow: worktree lifecycle becomes implicit in `next`/`complete`, not a separate concern
- tracking: Beads owns worktree CRUD; MindSpec coordinates with bead status

## ADR Touchpoints

- [ADR-0002](../../adr/ADR-0002.md): Beads as passive tracking substrate; MindSpec builds on top. Worktree CRUD is a Beads primitive; MindSpec orchestrates when it's called.
- [ADR-0003](../../adr/ADR-0003.md): MindSpec owns worktree conventions and orchestration rules. Beads provides the mechanism.
- [ADR-0005](../../adr/ADR-0005.md): State file tracks active bead; completion resets state. `mindspec complete` coordinates worktree cleanup with state advancement.
- [ADR-0007](../../adr/ADR-0007.md) (Proposed): Per-worktree state for parallel workstreams. Future extension point for spec/plan worktrees.

## Requirements

### Workflow integration (happy path)

1. **`mindspec next` creates worktree** — After claiming a bead and setting state to implement, `mindspec next` calls `bd worktree create worktree-<bead-id> --branch bead/<bead-id>` to create the worktree with Beads redirect. Prints the worktree path and instructs the agent to `cd` into it. If a worktree already exists for the bead (checked via `bd worktree list`), prints the existing path instead of creating a duplicate.
2. **`mindspec complete [bead-id]`** — New command that orchestrates the full bead close-out:
   - Validates all changes are committed (clean worktree via `git status --porcelain`)
   - Closes the bead (`bd close <id>`)
   - Removes the worktree (`bd worktree remove worktree-<bead-id>`)
   - Advances state (next bead, back to plan, or idle per ADR-0005)
   - Reports what was done
   - The bead ID defaults to the `activeBead` from state if not provided.
3. **`implement.md` template update** — The completion checklist references `mindspec complete` as the single close-out step, replacing the current multi-step checklist.

### Deprecation

4. **Deprecate `mindspec bead worktree`** — The existing `beadWorktreeCmd` is replaced by `bd worktree` for inspection and `mindspec next`/`mindspec complete` for the happy path. The command can print a deprecation notice pointing to `bd worktree list` and `mindspec complete`.
5. **Replace `internal/bead/worktree.go`** — Remove the custom `git worktree list --porcelain` parsing and `git worktree add` calls. Replace with `bd worktree` invocations where needed (e.g., in `mindspec next` for creation, `mindspec complete` for removal). Retain only the minimal helpers needed for worktree existence checks.

### Conventions

6. **Naming convention** — Worktree path: `../worktree-<bead-id>` (sibling to project root). Branch: `bead/<bead-id>`. Passed as arguments to `bd worktree create`.
7. **Beads daemon** — Worktrees should use `--no-daemon` mode or the sync-branch feature. Document this requirement. Consider setting `BEADS_NO_DAEMON=true` when creating worktrees.

## Scope

### In Scope
- `cmd/mindspec/complete.go` — new `mindspec complete` command
- `cmd/mindspec/next.go` — modify to create worktree via `bd worktree create` after claiming bead
- `internal/bead/worktree.go` — simplify to use `bd worktree` commands, remove custom git parsing
- `internal/bead/bdcli.go` — add `WorktreeCreate()`, `WorktreeList()`, `WorktreeRemove()` wrappers around `bd worktree`
- `internal/instruct/worktree.go` — simplify to use `bd worktree info`
- `cmd/mindspec/bead.go` — deprecate `beadWorktreeCmd`
- `internal/instruct/templates/implement.md` — update completion checklist
- `cmd/mindspec/root.go` — register `complete` command
- Updates to `CLAUDE.md`, `docs/core/CONVENTIONS.md`

### Out of Scope
- Worktrees for spec/plan phases (pending ADR-0007)
- Changes to Beads CLI itself
- Multi-repo worktree management

## Non-Goals

- Building a standalone `mindspec worktree` command tree — Beads already provides `bd worktree` and users can use it directly
- Worktree-per-spec or worktree-per-plan (pending ADR-0007; only worktree-per-bead in v1)
- Branch protection / PR-based merging to main (see ADR-0006, Proposed)

## Acceptance Criteria

### Workflow integration
- [ ] `mindspec next` creates a worktree via `bd worktree create worktree-<bead-id> --branch bead/<bead-id>` and prints the path
- [ ] `mindspec next` reuses an existing worktree if one already exists for the claimed bead
- [ ] `mindspec complete` closes the bead, removes the worktree via `bd worktree remove`, and advances state
- [ ] `mindspec complete` with no argument uses the `activeBead` from state
- [ ] `mindspec complete` refuses if the worktree has uncommitted changes, with exit code 1
- [ ] The `implement.md` instruction template references `mindspec complete` as the single close-out step

### Beads integration
- [ ] Worktree creation uses `bd worktree create` (not raw `git worktree add`), ensuring Beads redirect is set up
- [ ] Worktree removal uses `bd worktree remove` (not raw `git worktree remove`)
- [ ] `bd worktree list` shows worktrees created by `mindspec next`
- [ ] `bd worktree info` works in worktrees created by `mindspec next`

### Deprecation
- [ ] `mindspec bead worktree` prints a deprecation notice
- [ ] `internal/bead/worktree.go` no longer contains custom `git worktree list --porcelain` parsing

### General
- [ ] All new code has unit tests; `make test` passes
- [ ] Doc-sync: CLAUDE.md and CONVENTIONS.md updated

## Validation Proofs

- `./bin/mindspec next`: Claims a bead, creates worktree via `bd worktree create`, prints path
- `bd worktree list`: Shows the worktree created by `mindspec next`
- `./bin/mindspec complete`: Closes bead, removes worktree via `bd worktree remove`, advances state
- `make test`: All tests pass

## Open Questions

None — all resolved.

### Resolved

- ~~Should worktrees be created implicitly by workflow commands?~~ **Resolved**: Yes. `mindspec next` creates, `mindspec complete` cleans. No standalone `mindspec worktree` commands.
- ~~Should MindSpec implement its own worktree CRUD?~~ **Resolved**: No. Delegate to `bd worktree` commands. MindSpec provides orchestration only.
- ~~What about parallel spec/plan work?~~ **Resolved**: Deferred to ADR-0007 (Proposed). v1 scope is implementation beads only.
- ~~Should MindSpec state be stored in Beads?~~ **Resolved**: No. Beads state (issues) is shared across worktrees by design; MindSpec state (mode) needs to be per-worktree. Different concerns, different stores. See ADR-0007.

## Approval

- **Status**: DRAFT
- **Approved By**: —
- **Approval Date**: —
- **Notes**: Revised to delegate worktree CRUD to Beads and focus MindSpec on workflow orchestration only.
