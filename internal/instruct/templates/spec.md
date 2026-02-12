# MindSpec — Spec Mode

**Active Spec**: `{{.ActiveSpec}}`
{{- if .SpecGoal}}
**Goal**: {{.SpecGoal}}
{{- end}}

## Objective

Discuss user-facing value and define what "done" means. Spec Mode is intentionally implementation-light.

## Permitted Actions

- Create/edit `docs/specs/{{.ActiveSpec}}/spec.md`
- Create/edit domain docs (`docs/domains/`)
- Add glossary entries (`GLOSSARY.md`)
- Edit architecture docs (`docs/core/`)
- Draft ADRs (`docs/adr/`)

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

- **Spec approval**: Use `/spec-approve` when the spec is ready for planning

## Next Action

Complete the spec at `docs/specs/{{.ActiveSpec}}/spec.md`, then run `/spec-approve`.
