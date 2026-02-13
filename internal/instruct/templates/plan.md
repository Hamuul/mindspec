# MindSpec — Plan Mode

**Active Spec**: `{{.ActiveSpec}}`
{{- if .SpecGoal}}
**Goal**: {{.SpecGoal}}
{{- end}}

## Objective

Turn the approved spec into bounded, executable work chunks (implementation beads).

## Required Review (before planning)

1. Read accepted ADRs for impacted domains
2. Read domain docs (`overview.md`, `architecture.md`, `interfaces.md`)
3. Check Context Map (`docs/context-map.md`) for neighboring context contracts
4. Verify existing constraints and invariants

## Permitted Actions

- Create/edit `docs/specs/{{.ActiveSpec}}/plan.md`
- Create implementation beads in Beads (`bd create`)
- Propose new ADRs if divergence detected
- Update documentation to clarify scope

## Forbidden Actions

- Writing implementation code (`cmd/`, `internal/`, or equivalent)
- Writing test code
- Widening scope beyond the spec's defined user value

## Required Output

Implementation beads, each with:
- Small scope (one slice of value)
- 3-7 step micro-plan
- Explicit verification steps
- Dependencies between beads
- ADR citations

## Human Gates

- **Plan approval**: Use `/plan-approve` when the plan is ready
- **ADR divergence**: If an accepted ADR blocks progress, stop and inform the user

## Next Action

Complete the plan at `docs/specs/{{.ActiveSpec}}/plan.md`, then run `/plan-approve`.

## Session Close

Before ending a session: commit all changes, run quality gates if code changed, update bead status, and push to remote (if configured). Work is not complete until changes are committed and pushed.
