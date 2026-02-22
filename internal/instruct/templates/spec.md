# MindSpec — Spec Mode

**Active Spec**: `{{.ActiveSpec}}`
{{- if .SpecGoal}}
**Goal**: {{.SpecGoal}}
{{- end}}

## Objective

Discuss user-facing value and define what "done" means. Spec Mode is intentionally implementation-light.

## Permitted Actions

- Create/edit `.mindspec/docs/specs/{{.ActiveSpec}}/spec.md`
- Create/edit domain docs (`.mindspec/docs/domains/`)
- Add glossary entries (`GLOSSARY.md`)
- Edit architecture docs (`.mindspec/docs/core/`)
- Draft ADRs (`.mindspec/docs/adr/`)

## Forbidden Actions

- Creating or modifying code (`cmd/`, `internal/`, or equivalent)
- Creating or modifying test code
- Changing build/config that affects runtime behavior

## Required Output

A spec containing:
- Problem statement and target user outcome
- Acceptance criteria (specific, measurable)
- Impacted domains and ADR touchpoints
- Non-goals / constraints
- All open questions resolved

## Human Gates

- **Spec approval**: You MUST run `mindspec approve spec {{.ActiveSpec}}` before starting any plan work. This gate resolves the spec-approve step in the lifecycle molecule. Skipping it causes mode resolution to remain stuck in spec mode.

## Next Action

Complete the spec at `.mindspec/docs/specs/{{.ActiveSpec}}/spec.md`, then run `mindspec approve spec {{.ActiveSpec}}`. Do NOT create `plan.md` or begin planning until this command succeeds.

## Session Close

Before ending a session: commit all changes, run quality gates if code changed, update bead status, and push to remote (if configured). Work is not complete until changes are committed and pushed.
