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
5. **ADR Fitness Evaluation**: After reviewing ADRs, actively evaluate whether each relevant ADR still represents the best architectural choice for the work being planned. Do not blindly conform — if a better design would diverge from an accepted ADR, propose the divergence with justification. Prefer adherence when ADRs are sound; propose superseding when they are not. Document your evaluation in the `## ADR Fitness` section of the plan.

## Permitted Actions

- Create/edit `docs/specs/{{.ActiveSpec}}/plan.md`
- Create implementation beads in Beads (`bd create`)
- Propose new ADRs if divergence detected (`mindspec adr create --supersedes <old-id>`)
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

ADR Fitness evaluation (`## ADR Fitness` section in plan.md)

## Human Gates

- **Plan approval**: Use `/plan-approve` when the plan is ready
- **ADR divergence**: If a better design would diverge from an accepted ADR, **stop planning**. Present: (1) which ADR, (2) why it should be superseded, (3) the proposed alternative. Wait for human approval before proceeding. Use `mindspec adr create --supersedes <ADR-NNNN>` to create the superseding ADR once approved.

## Next Action
{{- if .PlanApproved}}

Plan is approved. Run `mindspec next` to claim the first bead and enter Implementation Mode. Do NOT manually set state to implement — `mindspec next` handles bead selection and state transition together.
{{- else}}

Complete the plan at `docs/specs/{{.ActiveSpec}}/plan.md`, then run `/plan-approve`.
{{- end}}

## Session Close

Before ending a session: commit all changes, run quality gates if code changed, update bead status, and push to remote (if configured). Work is not complete until changes are committed and pushed.
