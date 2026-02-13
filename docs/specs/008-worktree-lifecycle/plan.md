---
status: Draft
spec_id: 008-worktree-lifecycle
version: "1.0"
last_updated: 2026-02-12
adr_citations:
  - id: ADR-0002
    sections: ["Parallelism Compatibility"]
  - id: ADR-0003
    sections: ["MindSpec owns worktree conventions"]
  - id: ADR-0005
    sections: ["Completion and Reset"]
work_chunks:
  - id: 1
    title: "Core worktree package extraction"
    scope: "internal/worktree/worktree.go, internal/worktree/worktree_test.go, internal/bead/worktree.go"
    verify:
      - "ParseList() parses multi-worktree porcelain output correctly"
      - "FindByBead() matches by path suffix or branch convention"
      - "Create() validates in_progress status and clean tree"
      - "internal/bead/worktree.go delegates to internal/worktree/"
      - "make test passes"
    depends_on: []
  - id: 2
    title: "Worktree list with bead enrichment"
    scope: "internal/worktree/list.go, internal/worktree/list_test.go"
    verify:
      - "EnrichedList() extracts bead ID from path/branch naming convention"
      - "EnrichedList() looks up bead status via bd show"
      - "Worktrees with no matching bead are flagged as orphaned"
      - "FormatList() produces readable tabular output"
      - "make test passes"
    depends_on: [1]
  - id: 3
    title: "Worktree cleanup"
    scope: "internal/worktree/clean.go, internal/worktree/clean_test.go"
    verify:
      - "Clean() refuses if bead is not closed"
      - "Clean() refuses if worktree has uncommitted changes"
      - "Clean() runs git worktree remove + git branch -d"
      - "CleanAll() iterates worktrees and cleans only those with closed beads"
      - "CleanAll() is dry-run by default; only executes when dryRun=false"
      - "make test passes"
    depends_on: [1]
  - id: 4
    title: "CLI wiring + template update + doc-sync"
    scope: "cmd/mindspec/worktree.go, cmd/mindspec/root.go, internal/instruct/templates/implement.md, CLAUDE.md, docs/core/CONVENTIONS.md"
    verify:
      - "./bin/mindspec worktree --help shows create, list, clean subcommands"
      - "Exit codes: 0 success, 1 validation, 2 git/bd error"
      - "implement.md template includes worktree clean in completion checklist"
      - "CLAUDE.md and CONVENTIONS.md updated"
      - "mindspec bead worktree still works (backward compatibility)"
      - "make build && make test passes"
    depends_on: [2, 3]
---

# Plan: Spec 008 — Worktree Lifecycle Management

**Spec**: [spec.md](spec.md)

---

## Design Notes

### Package Architecture

New `internal/worktree/` package owns all git worktree operations. This extracts and extends logic currently in `internal/bead/worktree.go`.

The `internal/bead/worktree.go` file becomes a thin delegation layer — its exported functions (`ParseWorktreeList`, `FindWorktree`, `CreateWorktree`) call through to `internal/worktree/` to preserve backward compatibility for `mindspec bead worktree`.

Import direction: `internal/bead/` → `internal/worktree/` (bead depends on worktree, not the reverse). The worktree package has no dependency on bead.

### Bead Status Lookups

The worktree package needs bead status for list enrichment and clean validation. Rather than importing `internal/bead/` (which would create a circular dependency), the worktree package shells out to `bd show <id> --json` directly through its own `execCommand` variable. This is the same pattern used in `internal/bead/bdcli.go` and keeps the package self-contained.

### execCommand Pattern

Same testability pattern as `internal/bead/`:

```go
var execCommand = exec.Command
```

Tests override this to capture arguments or return canned output. The worktree package needs both `git` and `bd` command execution.

### Enriched List Model

`ListEntry` extends `WorktreeEntry` with bead-specific metadata:

```go
type ListEntry struct {
    WorktreeEntry
    BeadID     string // extracted from path/branch convention
    BeadStatus string // from bd show, or "" if lookup fails
    IsOrphan   bool   // true if no bead ID could be extracted
    IsMain     bool   // true if this is the main worktree (not bead-associated)
}
```

The main worktree (project root) is always present in `git worktree list` output but isn't bead-associated — it gets `IsMain: true` and is excluded from orphan detection.

### Clean Safety

`Clean()` has three safety checks before removing:

1. **Bead is closed** — refuses if bead is open/in_progress (prevents accidental deletion of active work)
2. **Worktree is clean** — runs `git -C <path> status --porcelain` to check for uncommitted changes in the worktree itself
3. **Worktree exists** — `FindByBead()` must locate the worktree path

Removal sequence: `git worktree remove <path>` then `git branch -d bead/<bead-id>`. The `-d` flag (not `-D`) ensures the branch is only deleted if fully merged, providing an additional safety net.

### CleanAll Dry-Run

`CleanAll(root, dryRun)` iterates all worktrees from `EnrichedList()`, filters to those where `BeadStatus == "closed"`, and either reports what would be removed (dry-run) or executes removal. Returns a list of action strings for reporting.

### Exit Code Convention

Consistent with Spec 007:

- 0: success
- 1: validation failure (bead not in correct state, dirty tree, etc.)
- 2: git or bd CLI error

---

## Bead 008-1: Core worktree package extraction

**Scope**: `internal/worktree/worktree.go`, `internal/worktree/worktree_test.go`, `internal/bead/worktree.go`

**Steps**:
1. Create `internal/worktree/worktree.go` with `var execCommand`, `WorktreeEntry` struct, `ParseList()`, `parseWorktreePorcelain()` — extracted from `internal/bead/worktree.go`
2. Add `FindByBead(beadID string) (string, error)` — same logic as current `FindWorktree()`
3. Add `Create(root, beadID string) (string, error)` — same logic as current `CreateWorktree()`, but calls `bd show` directly for bead status validation
4. Add `CheckCleanTree(dir string) error` — generalized to accept a directory (worktree path or project root)
5. Update `internal/bead/worktree.go` to import `internal/worktree/` and delegate all exported functions
6. Write tests: porcelain parsing, find by path/branch, create argument construction, clean tree check, delegation from bead package

**Verification**:
- [ ] `ParseList()` parses multi-worktree porcelain output correctly
- [ ] `FindByBead()` matches by path suffix or branch convention
- [ ] `Create()` validates `in_progress` status and clean tree
- [ ] `CheckCleanTree()` works for both project root and worktree paths
- [ ] `internal/bead/worktree.go` delegates to `internal/worktree/` without behavior change
- [ ] `make test` passes

**Depends on**: nothing

---

## Bead 008-2: Worktree list with bead enrichment

**Scope**: `internal/worktree/list.go`, `internal/worktree/list_test.go`

**Steps**:
1. Define `ListEntry` struct (embeds `WorktreeEntry`, adds `BeadID`, `BeadStatus`, `IsOrphan`, `IsMain`)
2. Implement `extractBeadID(entry WorktreeEntry) string` — extracts from `worktree-<id>` path suffix or `bead/<id>` branch
3. Implement `lookupBeadStatus(beadID string) string` — calls `bd show <id> --json`, returns status or `"unknown"` on error
4. Implement `EnrichedList() ([]ListEntry, error)` — calls `ParseList()`, enriches each entry, flags orphans (non-main entries with no bead ID)
5. Implement `FormatList(entries []ListEntry) string` — tabular output: path, branch, bead ID, status, flags
6. Write tests: bead ID extraction from path and branch, orphan detection, main worktree exclusion, format output

**Verification**:
- [ ] `EnrichedList()` extracts bead ID from path/branch naming convention
- [ ] `EnrichedList()` looks up bead status via `bd show`
- [ ] Worktrees with no matching bead are flagged as orphaned
- [ ] Main worktree is not flagged as orphaned
- [ ] `FormatList()` produces readable tabular output
- [ ] `make test` passes

**Depends on**: 008-1

---

## Bead 008-3: Worktree cleanup

**Scope**: `internal/worktree/clean.go`, `internal/worktree/clean_test.go`

**Steps**:
1. Implement `Clean(root, beadID string) error` — find worktree via `FindByBead()`, validate bead closed via `bd show`, check worktree clean via `CheckCleanTree(wtPath)`, run `git worktree remove <path>`, run `git branch -d bead/<beadID>`
2. Implement `CleanAll(root string, dryRun bool) ([]string, error)` — call `EnrichedList()`, filter to `BeadStatus == "closed"`, iterate and clean (or report if dry-run), return action descriptions
3. Handle edge cases: worktree not found (already removed), branch not found (already deleted), bead lookup failure (skip with warning)
4. Write tests: refuses non-closed bead, refuses dirty worktree, correct git commands for removal, dry-run returns descriptions without executing, already-removed worktree handled gracefully

**Verification**:
- [ ] `Clean()` refuses if bead is not closed
- [ ] `Clean()` refuses if worktree has uncommitted changes
- [ ] `Clean()` runs `git worktree remove` + `git branch -d`
- [ ] `CleanAll()` iterates worktrees and cleans only those with closed beads
- [ ] `CleanAll()` is dry-run by default; only executes when `dryRun=false`
- [ ] Edge cases handled gracefully (missing worktree, missing branch)
- [ ] `make test` passes

**Depends on**: 008-1 (uses `FindByBead`, `CheckCleanTree`); also uses `EnrichedList` from 008-2

---

## Bead 008-4: CLI wiring + template update + doc-sync

**Scope**: `cmd/mindspec/worktree.go`, `cmd/mindspec/root.go`, `internal/instruct/templates/implement.md`, `CLAUDE.md`, `docs/core/CONVENTIONS.md`

**Steps**:
1. Create `cmd/mindspec/worktree.go`: parent `worktreeCmd` + three child commands, following `bead.go` pattern
2. Wire `worktreeCreateCmd` (ExactArgs(1)): findRoot → Preflight → `worktree.Create()` → print path
3. Wire `worktreeListCmd` (NoArgs): findRoot → Preflight → `worktree.EnrichedList()` → `worktree.FormatList()` → print
4. Wire `worktreeCleanCmd` (0 or 1 args, `--all`, `--yes` flags): findRoot → Preflight → if `--all`: `CleanAll(root, !yes)`, else `Clean(root, args[0])`
5. Register `worktreeCmd` in `root.go` init()
6. Update `internal/instruct/templates/implement.md`: add `mindspec worktree clean <bead-id>` to the completion checklist between bead closure and state advancement
7. Update `CLAUDE.md`: add `mindspec worktree` commands to Build & Run and command table
8. Update `docs/core/CONVENTIONS.md`: document worktree lifecycle (create on bead start, clean on bead close)

**Verification**:
- [ ] `./bin/mindspec worktree --help` shows `create`, `list`, `clean` subcommands
- [ ] Exit codes: 0 success, 1 validation, 2 git/bd error
- [ ] `implement.md` template includes `mindspec worktree clean` in completion checklist
- [ ] `CLAUDE.md` and `CONVENTIONS.md` updated
- [ ] `mindspec bead worktree` still works (backward compatibility)
- [ ] `make build && make test` passes

**Depends on**: 008-2, 008-3

---

## Dependency Graph

```
008-1 (core worktree package)
  ├── 008-2 (list with bead enrichment)
  └── 008-3 (worktree cleanup)
        ↓ both
      008-4 (CLI wiring + template + doc-sync)
```

008-2 and 008-3 are parallelizable after 008-1. Note: 008-3's `CleanAll` uses `EnrichedList` from 008-2, so while they can be developed in parallel, 008-3 has a soft dependency on 008-2's `EnrichedList()` being available.

---

## End-to-End Verification

```bash
make build && make test
./bin/mindspec worktree --help
./bin/mindspec worktree list
./bin/mindspec worktree clean --all         # dry-run
./bin/mindspec worktree clean --all --yes   # execute
./bin/mindspec bead worktree <id>           # backward compat
```
