---
status: Draft
spec_id: "057-worktree-aware-resolution"
version: 1
last_updated: "2026-03-01"
approved_at: ""
approved_by: ""
bead_ids: []
adr_citations:
  - id: ADR-0006
    sections: ["Branch topology", "Worktree lifecycle"]
  - id: ADR-0019
    sections: ["Enforcement layers"]
  - id: ADR-0022
    sections: ["Spec artifact resolution", "ActiveSpecs worktree-aware"]
---

# Plan: 057-worktree-aware-resolution

## ADR Fitness

- **ADR-0006** (Protected Main, PR-based merging): Sound. This spec implements the resolution side of worktree isolation that ADR-0006 assumes. No divergence.
- **ADR-0019** (Deterministic enforcement): Sound. Enforcement layers remain unchanged; this spec fixes the path resolution that enforcement depends on. No divergence.
- **ADR-0022** (Worktree-aware spec resolution): This spec is the direct implementation of ADR-0022. The resolution algorithm (check worktree → main → legacy) matches exactly. No divergence.

## Testing Strategy

- **Unit tests**: `internal/workspace/workspace_test.go` — new tests for worktree-aware `SpecDir`, updated `EffectiveSpecRoot` tests become `SpecDir` tests
- **Package tests**: `go test ./internal/complete/ ./internal/resolve/ ./internal/next/ ./internal/approve/ ./cmd/mindspec/ -short` — all affected packages pass
- **Vet**: `go vet ./...` clean
- **LLM integration**: `TestLLM_CompleteFromSpecWorktree` continues to pass

## Bead 1: Make SpecDir worktree-aware and remove EffectiveSpecRoot

**Steps**
1. Refactor `workspace.SpecDir(root, specID)` to check `root/.worktrees/worktree-spec-<specID>/.mindspec/docs/specs/<specID>/` first, then `root/.mindspec/docs/specs/<specID>/`, then `root/docs/specs/<specID>/` (legacy fallback)
2. Update `LifecyclePath(root, specID)` and `RecordingDir(root, specID)` — these already delegate to `SpecDir`, so they inherit worktree-awareness automatically (verify)
3. Remove `EffectiveSpecRoot` function (or mark deprecated with panic if called)
4. Update `workspace_test.go`: convert `TestEffectiveSpecRoot_*` tests to `TestSpecDir_WorktreeAware_*` tests that validate the 3-step resolution order

**Verification**
- [ ] `go test ./internal/workspace/ -v -run TestSpecDir` passes
- [ ] `go vet ./internal/workspace/` clean

**Depends on**
None

## Bead 2: Update all production callers to use plain SpecDir

**Steps**
1. `internal/approve/spec.go`: Remove `effectiveRoot := workspace.EffectiveSpecRoot(...)`, use `workspace.SpecDir(root, specID)` directly
2. `internal/approve/plan.go` (lines 42, 329): Same pattern — remove `effectiveRoot`, use `workspace.SpecDir(root, specID)` directly
3. `internal/complete/complete.go` (`advanceState`): Replace `effectiveRoot + manual filepath.Join` with `workspace.SpecDir(root, specID)`
4. `internal/next/beads.go` (`ResolveActiveBead`): Replace `effectiveRoot + manual filepath.Join` with `workspace.SpecDir(root, specID)`
5. `cmd/mindspec/validate.go`: Replace both `effectiveRoot` usages with direct `workspace.SpecDir(root, specID)` passed to validators
6. `internal/lifecycle/scenario_test.go`: Update test code callers (replace `EffectiveSpecRoot` with `SpecDir`)

**Verification**
- [ ] `go build ./...` compiles (no references to removed `EffectiveSpecRoot`)
- [ ] `go test ./internal/approve/ ./internal/complete/ ./internal/next/ ./cmd/mindspec/ -short` passes
- [ ] `go vet ./...` clean
- [ ] Zero `grep -r EffectiveSpecRoot` matches in `.go` files (excluding test history/docs)

**Depends on**
Bead 1

## Bead 3: Make ActiveSpecs worktree-aware

**Steps**
1. In `resolve.ActiveSpecs(root)`: after scanning `DocsDir(root)/specs/*/lifecycle.yaml`, also scan `root/.worktrees/worktree-spec-*/` directories for lifecycle.yaml
2. Deduplicate by specID (worktree result wins over main repo if both exist)
3. Keep the sort-by-specID behavior
4. Update `ResolveMode(root, specID)` to use `workspace.SpecDir(root, specID)` instead of manual `DocsDir(root)/specs/<specID>` path construction

**Verification**
- [ ] `go test ./internal/resolve/ -v` passes (existing tests use main repo layout, still works via fallback)
- [ ] `go test ./internal/resolve/ -v -run TestActiveSpecs` — verify worktree scanning doesn't break existing behavior
- [ ] `go vet ./internal/resolve/` clean

**Depends on**
Bead 1

## Bead 4: Final validation and cleanup

**Steps**
1. Run full test suite: `make test`
2. Run `go vet ./...`
3. Run LLM test: `env -u CLAUDECODE go test ./internal/harness/ -v -run TestLLM_CompleteFromSpecWorktree -timeout 10m`
4. Verify zero manual spec path construction in production code outside workspace package
5. Remove any remaining dead code or comments referencing `EffectiveSpecRoot`

**Verification**
- [ ] `make test` passes
- [ ] `go vet ./...` clean
- [ ] LLM test `CompleteFromSpecWorktree` passes
- [ ] `grep -rn 'EffectiveSpecRoot' --include='*.go'` returns only test files and docs

**Depends on**
Bead 2, Bead 3

## Provenance

| Acceptance Criterion | Bead | Verification |
|:-|:-|:-|
| `SpecDir` returns worktree path when spec worktree exists | Bead 1 | `TestSpecDir_WorktreeAware` |
| `SpecDir` returns main repo path when no worktree | Bead 1 | `TestSpecDir_WorktreeAware` |
| `SpecDir` returns canonical path for new spec creation | Bead 1 | `TestSpecDir_WorktreeAware` |
| `ActiveSpecs` finds specs only in worktrees | Bead 3 | `TestActiveSpecs` |
| Zero production callers of `EffectiveSpecRoot` | Bead 2, 4 | grep validation |
| Zero manual spec path construction outside workspace | Bead 2, 4 | grep validation |
| All existing unit tests pass | Bead 4 | `make test` |
| LLM test passes | Bead 4 | `TestLLM_CompleteFromSpecWorktree` |
| `go vet ./...` clean | Bead 4 | `go vet` |
