---
adr_citations:
    - id: ADR-0006
      sections:
        - ADR Fitness
    - id: ADR-0019
      sections:
        - ADR Fitness
last_updated: "2026-02-26T12:00:00Z"
spec_id: 047-clean-context-enforcement
status: Draft
version: 1
---

# Plan: 047-clean-context-enforcement — Clean Context Enforcement for Bead Starts

## Overview

When an agent implements multiple beads sequentially in one session, stale context from prior beads degrades output quality. This plan adds a clear gate (`needs_clear` flag) that blocks `mindspec next` after `mindspec complete`, forcing a context reset before the next bead starts. It also adds a bead-specific context primer that replaces generic implement-mode guidance with focused, bead-scoped information.

The implementation has four beads:

1. **State + clear gate** (Bead 1): `needs_clear` flag, gate logic in complete and next
2. **Bead context primer** (Bead 2): Primer builder, template, token estimation
3. **Multi-agent emit-only** (Bead 3): `--emit-only` flag for team lead usage
4. **Hook enforcement** (Bead 4): PreToolUse hook, SessionStart hook update

## ADR Fitness

- **ADR-0006** (Protected main with PR-based merging): Sound. This spec does not change branching or merge behavior. The clear gate operates entirely within the existing worktree lifecycle — it gates context, not code flow. No divergence.
- **ADR-0019** (Deterministic worktree enforcement): Sound. This spec extends the enforcement pattern from ADR-0019 (three layers: git hook, CLI guard, agent hook) to a new concern — context hygiene. The PreToolUse hook (R4) follows the same exit-code-2 blocking pattern established for worktree enforcement. No divergence.

## Testing Strategy

- **Unit tests**: Each bead includes tests for its package. State flag round-trip, gate logic (block/force/skip), primer rendering, emit-only path.
- **Integration**: `make test` passes after each bead. `make build` succeeds.
- **Manual verification**: Complete a bead with another ready → `needs_clear` set → `mindspec next` blocked → `/clear` + session restart → `mindspec next` succeeds with bead primer.

## Bead 1: State flag and clear gate logic

**Provenance**: R1 (Single-agent clear gate), R5 (Graceful degradation / --force)

**Steps**
1. Add `NeedsClear bool` field to `State` struct in `internal/state/state.go` with JSON tag `"needs_clear,omitempty"`. Using `omitempty` keeps state.json clean when the flag is false.
2. Add `ClearNeedsClear(root string) error` helper in `internal/state/state.go` — reads state, sets `NeedsClear = false`, writes state. This is the function the SessionStart hook will call.
3. Add `state clear-flag` subcommand in `cmd/mindspec/state.go` that calls `state.ClearNeedsClear(root)`. This gives the SessionStart hook a CLI entry point.
4. Modify `internal/complete/complete.go` `Run()`: after the `case state.ModeImplement:` branch in the state advance switch (line ~146), read state back, set `NeedsClear = true`, and write. This must come after `setModeFn` since `setModeFn` constructs a fresh `State{}`.
5. Modify `cmd/mindspec/next.go` RunE: after finding root and before the clean tree check, read state. If `NeedsClear` is true and `--force` is not set, exit with error: `"Context clear required. Run /clear to reset your context, then retry.\nUse --force to bypass."`. If `--force` is set, print a warning and continue.
6. Add `--force` flag to `nextCmd` in `init()`.
7. Add unit tests:
   - `internal/state/state_test.go`: `NeedsClear` round-trip (set true, read back, clear, read back)
   - `internal/complete/complete_test.go`: verify `NeedsClear` is set when `advanceState` returns `ModeImplement`
   - Test that `NeedsClear` is NOT set when advancing to review/idle/plan

**Verification**
- [ ] `go test ./internal/state/...` passes with NeedsClear round-trip
- [ ] `go test ./internal/complete/...` passes with NeedsClear assertion
- [ ] `mindspec complete` on a bead with a successor → `state.json` shows `"needs_clear": true`
- [ ] `mindspec next` with `needs_clear: true` → exits with error
- [ ] `mindspec next --force` with `needs_clear: true` → proceeds with warning
- [ ] `make test` passes

**Depends on**
None

## Bead 2: Bead context primer

**Provenance**: R2 (Bead context primer), R6 (Context budget estimation)

**Steps**
1. Create `internal/instruct/primer.go` with:
   - `BeadPrimerContext` struct: `BeadID`, `BeadTitle`, `BeadDescription`, `SpecID`, `SpecSlice` (requirements + acceptance criteria from spec.md), `PlanSlice` (the matching `## Bead ...` section from plan.md), `FilePaths []string`, `ADRCitations []string`, `EstimatedTokens int`
   - `BuildBeadPrimer(root, specID, beadID string) (*BeadPrimerContext, error)`:
     a. Call `bead.Show(beadID)` to get title and description
     b. Read `spec.md` and extract `## Requirements` and `## Acceptance Criteria` sections (simple string scanning — find header, collect until next `## ` header)
     c. Read `plan.md` and extract the `## Bead ...` section matching the bead ID or title (scan for `## Bead` headers, match by bead ID substring)
     d. Parse plan.md YAML frontmatter for `adr_citations` (reuse the `yaml.v3` approach from `internal/validate/plan.go`)
     e. Extract file paths from the plan slice (scan for lines containing paths like `internal/`, `cmd/`, etc.)
     f. Render the primer template, then compute `trace.EstimateTokens()` on the rendered output
   - `RenderBeadPrimer(ctx *BeadPrimerContext) (string, error)` — executes the bead-primer template
2. Create `internal/instruct/templates/bead-primer.md` template:
   ```
   # Bead Context Primer
   **Spec**: {{.SpecID}} | **Bead**: {{.BeadID}} | **~{{.EstimatedTokens}} tokens**
   ## {{.BeadTitle}}
   {{.BeadDescription}}
   ## Spec Slice
   {{.SpecSlice}}
   ## Plan Work Chunk
   {{.PlanSlice}}
   ## Key File Paths
   {{range .FilePaths}}- {{.}}
   {{end}}
   ## ADR Citations
   {{range .ADRCitations}}- {{.}}
   {{end}}
   ```
3. Modify `cmd/mindspec/next.go`: after claiming the bead and updating state (step 7), replace the generic `emitInstruct(root)` call with primer output when mode is implement:
   - Call `instruct.BuildBeadPrimer(root, specID, beadID)`
   - If successful, call `instruct.RenderBeadPrimer(ctx)` and print to stdout
   - If primer fails (graceful degradation), fall back to `emitInstruct(root)`
4. Add unit tests:
   - `internal/instruct/primer_test.go`: test `BuildBeadPrimer` with mock spec.md and plan.md on disk, verify section extraction, token estimate > 0
   - Test graceful degradation when spec.md or plan.md is missing

**Verification**
- [ ] `go test ./internal/instruct/...` passes with primer tests
- [ ] `mindspec next` emits a bead-specific context primer (bead description, spec slice, plan slice, file paths, ADR citations)
- [ ] Primer includes estimated token count
- [ ] Primer gracefully degrades to generic instruct when spec/plan unavailable
- [ ] `make test` passes

**Depends on**
Bead 1

## Bead 3: Multi-agent emit-only mode

**Provenance**: R3 (Multi-agent context handoff)

**Steps**
1. Add `--emit-only` flag to `nextCmd` in `cmd/mindspec/next.go` `init()`.
2. In RunE, when `--emit-only` is set:
   - Skip the clear gate check (emit-only is for fresh agents that have no prior context)
   - Query ready beads as normal
   - Accept an optional positional argument as an explicit bead ID. If provided, use `bead.Show(id)` to fetch it instead of querying ready beads.
   - Skip `ClaimBead`, `EnsureWorktree`, state update, and recording
   - Build and render the bead primer
   - Print primer to stdout and return
3. Add `FetchBeadByID(id string) (BeadInfo, error)` to `internal/next/beads.go` — calls `bd show <id> --json` and parses a single `BeadInfo`. This is needed for the explicit bead ID path.
4. Add unit tests:
   - Test emit-only path prints primer and does not claim bead
   - Test emit-only with explicit bead ID

**Verification**
- [ ] `mindspec next --emit-only` outputs primer to stdout, bead remains unclaimed
- [ ] `mindspec next --emit-only <bead-id>` outputs primer for the specified bead
- [ ] `go test ./internal/next/...` passes
- [ ] `make test` passes

**Depends on**
Bead 2

## Bead 4: Hook enforcement and SessionStart integration

**Provenance**: R4 (Hook enforcement for clear gate)

**Steps**
1. Extend `internal/setup/claude.go` `wantedHooks()`: add a `PreToolUse` entry with matcher `"Bash"`:
   - The hook command reads `.mindspec/state.json` via `jq`
   - Checks if `needs_clear` is `true`
   - Checks if the input command contains `mindspec next` but NOT `--force`
   - If both conditions met, exit 2 with message: `"needs_clear is set. Run /clear to reset your context, then retry mindspec next. Use --force to bypass."`
   - If `needs_clear` is false or command doesn't match, exit 0
2. Update the SessionStart hook command in `internal/setup/claude.go`: prepend `mindspec state clear-flag 2>/dev/null;` before the existing `mindspec instruct` call. This clears the flag on every session start (which happens after `/clear`).
3. Run `mindspec setup claude` to install the updated hooks (document this in verification).
4. Add unit tests:
   - `internal/setup/claude_test.go`: verify `wantedHooks()` includes the new Bash PreToolUse hook
   - Verify SessionStart command includes `state clear-flag`

**Verification**
- [ ] `mindspec setup claude` installs the PreToolUse Bash hook
- [ ] PreToolUse hook blocks `mindspec next` when `needs_clear` is true
- [ ] PreToolUse hook does NOT block `mindspec next --force`
- [ ] PreToolUse hook does NOT block other Bash commands when `needs_clear` is true
- [ ] SessionStart hook clears `needs_clear` flag on session start
- [ ] After `/clear` + session restart, `needs_clear` is false and `mindspec next` proceeds
- [ ] `make test` passes
- [ ] All existing tests pass (`make test`)
- [ ] New unit tests cover clear gate logic and primer generation

**Depends on**
Bead 1

## Dependency Graph

```
Bead 1 (state flag + clear gate)
  ├── Bead 2 (bead context primer)
  │     └── Bead 3 (emit-only mode)
  └── Bead 4 (hook enforcement)
```

Beads 2 and 4 can be worked in parallel since they share only Bead 1 as a dependency. Bead 3 depends on Bead 2 (needs the primer builder).

## Provenance

| Acceptance Criterion | Bead | Verification |
|:---------------------|:-----|:-------------|
| `complete` sets `needs_clear: true` when next bead ready | Bead 1 | state.json shows flag after complete |
| `next` refuses when `needs_clear` set | Bead 1 | Exit with error and instruction |
| `next --force` bypasses clear gate | Bead 1 | Proceeds with warning |
| After `/clear` + SessionStart, `needs_clear` reset | Bead 4 | `state clear-flag` in SessionStart hook |
| `next` emits bead-specific primer | Bead 2 | Primer output includes description, spec slice, plan slice, paths, ADRs |
| `next --emit-only` outputs primer without claiming | Bead 3 | Bead remains unclaimed, primer on stdout |
| PreToolUse hook blocks `next` when `needs_clear` set | Bead 4 | Hook exits 2, agent sees instruction |
| Primer includes estimated token count | Bead 2 | Token count line in primer output |
| All existing tests pass | All | `make test` green after each bead |
| New unit tests cover gate + primer | Bead 1, 2 | Test files in state, complete, instruct packages |

## Risk Notes

- **`setModeFn` overwrites state**: `state.SetMode` constructs a fresh `State{}` without `NeedsClear`. Bead 1 must read state *after* `setModeFn` writes, then set the flag. The two-write sequence (SetMode → read → set flag → write) is safe because complete runs single-threaded.
- **Primer section extraction is best-effort**: If spec.md or plan.md has non-standard headers, the primer falls back gracefully to generic instruct output. This is acceptable for v1 — structured extraction can be improved later.
- **PreToolUse hook matching**: The Bash hook matches `mindspec next` by substring in the command input. This could false-positive on commands like `echo "mindspec next"` — acceptable because the hook only blocks when `needs_clear` is also true, which is a narrow window.
- **`state clear-flag` race**: The SessionStart hook clears the flag before `mindspec instruct` runs. If instruct somehow sets the flag (it doesn't), there could be a race. In practice this is safe because only `complete` sets the flag.
