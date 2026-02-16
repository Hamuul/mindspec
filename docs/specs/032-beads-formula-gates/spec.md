# Spec 032-beads-formula-gates: Beads Formula Gates

## Goal

Replace mindspec's broken manual gate creation (`bd create --type=gate`) with beads formulas — the intended mechanism for workflow orchestration with human approval gates. Each spec lifecycle becomes a formula molecule with proper gate steps, resolved via `bd gate approve`.

## Background

Mindspec models the spec lifecycle as a sequence of human-gated phases: spec → spec-approve → plan → plan-approve → implement → review. Currently, `internal/bead/gate.go` attempts to create standalone gate issues via `bd create --type=gate`, which fails because `gate` is not a valid standalone issue type in beads — it's a formula step primitive.

The current code has fallback logic ("legacy beads — proceeding without gate") that silently skips gate enforcement, meaning approval gates are effectively unenforced. This defeats the purpose of having human gates in the workflow.

Beads provides a formula system (`bd pour`) that creates molecules with proper gate steps. Steps with `type = "human"` and `[steps.gate]` blocks are first-class async coordination points resolved via `bd gate approve`. This is the intended API for what mindspec is trying to do.

## Impacted Domains

- **bead**: `internal/bead/gate.go`, `internal/bead/spec.go`, `internal/bead/plan.go` — replace manual gate creation with formula-based molecule instantiation
- **approve**: `internal/approve/spec.go`, `internal/approve/plan.go` — resolve gates via `bd gate approve` instead of `bd gate resolve` or `bd close`
- **specinit**: `internal/specinit/specinit.go` — pour the lifecycle formula when initializing a spec
- **state**: `internal/state/state.go` — molecule step progression may inform state transitions

## ADR Touchpoints

None currently. May warrant a new ADR if the formula approach changes how specs are tracked in beads (molecule per spec vs. standalone issues).

## Requirements

1. Define a `spec-lifecycle.formula.toml` in `.beads/formulas/` with steps: spec, spec-approve (human gate), plan, plan-approve (human gate), implement, review
2. `mindspec spec-init` pours the formula via `bd pour spec-lifecycle --var spec_id=<id>`, creating a molecule that tracks the full lifecycle
3. `mindspec approve spec` resolves the spec-approve gate step via `bd gate approve`
4. `mindspec approve plan` resolves the plan-approve gate step via `bd gate approve`
5. Gate enforcement is real: `mindspec approve plan` must fail if the spec-approve gate is not yet resolved
6. Remove `internal/bead/gate.go` functions that use `bd create --type=gate` and `bd gate resolve`
7. Existing specs without molecules continue to work (backward compatibility via the existing "no gate found" fallback)

## Scope

### In Scope
- `.beads/formulas/spec-lifecycle.formula.toml` — new formula definition
- `internal/bead/gate.go` — rewrite to use formula/molecule gate APIs
- `internal/bead/spec.go` — replace `CreateSpecBead()` gate creation with molecule instantiation
- `internal/bead/plan.go` — replace `CreatePlanBeads()` gate creation with molecule step progression
- `internal/approve/spec.go` — use `bd gate approve` for spec gate resolution
- `internal/approve/plan.go` — use `bd gate approve` for plan gate resolution
- `internal/specinit/specinit.go` — pour formula on spec init
- `cmd/mindspec/approve.go` — wire updated approve logic

### Out of Scope
- Timer or GitHub gates (only human gates needed for now)
- Formula aspects or cross-cutting concerns
- Changing the spec folder layout or state machine

## Non-Goals

- Implementing automated gate types (timer, gh:run, gh:pr) — only human gates are needed
- Changing the existing mindspec CLI UX — `spec-init`, `approve spec`, `approve plan` keep the same interface
- Migrating existing closed/completed specs to the formula model

## Acceptance Criteria

- [ ] `mindspec spec-init 999-test` creates a beads molecule via `bd pour` with 6 steps (spec, spec-approve, plan, plan-approve, implement, review)
- [ ] `bd mol show <molecule-id>` shows the spec-approve step with gate status `pending`
- [ ] `mindspec approve spec 999-test` resolves the spec-approve gate; `bd gate list` confirms it is closed
- [ ] `mindspec approve plan 999-test` fails with an error if spec-approve gate has not been resolved
- [ ] `mindspec approve plan 999-test` succeeds after spec-approve is resolved, resolving the plan-approve gate
- [ ] Running `mindspec approve spec` on a pre-032 spec (no molecule) still works via backward-compat fallback
- [ ] All existing tests in `internal/bead/`, `internal/approve/`, `internal/specinit/` pass
- [ ] `make test` passes with no new failures

## Validation Proofs

- `mindspec spec-init 999-test && bd mol list --json`: molecule exists with spec-lifecycle formula
- `mindspec approve spec 999-test && bd gate list --all --json`: spec-approve gate is closed
- `mindspec approve plan 999-test` (before spec approval): exits with non-zero status and error message
- `make test`: all tests pass

## Open Questions

- [ ] Does `bd pour` return the molecule ID in JSON output, and can we store it in `.mindspec/state.json` or the spec frontmatter for later lookup?
- [ ] Does `bd gate approve` accept a step ID within a molecule, or does it need the gate's issue ID? Need to verify the exact CLI interface.
- [ ] Should the formula be version-controlled in `.beads/formulas/` (project-level) or shipped as part of the mindspec binary (embedded)?

## Approval

- **Status**: DRAFT
- **Approved By**: —
- **Approval Date**: —
- **Notes**: —
