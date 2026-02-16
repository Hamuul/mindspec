# MindSpec — Review Mode

**Active Spec**: `{{.ActiveSpec}}`
{{- if .SpecGoal}}
**Goal**: {{.SpecGoal}}
{{- end}}

## Objective

All implementation beads are complete. Present the work for human review before closing out.

## Permitted Actions

- Reading code, tests, and documentation to verify completeness
- Running tests and quality gates
- Minor fixes discovered during review (typos, formatting)
- Updating documentation if gaps are found

## Forbidden Actions

- New feature work (create a new spec instead)
- Significant refactoring beyond the spec's scope
- Closing out without human approval

## Review Checklist

1. **Acceptance criteria**: Read the spec at `docs/specs/{{.ActiveSpec}}/spec.md` and verify each acceptance criterion is met
2. **Tests**: Run `make test` and confirm all tests pass
3. **Build**: Run `make build` and confirm clean build
4. **Doc-sync**: Verify documentation matches the implementation
5. **Summary**: Present a brief summary of what was built and how each acceptance criterion was satisfied

## Human Gate

- **Implementation approval**: Run `mindspec approve impl <id>` when the human accepts the implementation

## Next Action

Read the spec's acceptance criteria, verify each one, and present the review summary to the human. When they approve, run `mindspec approve impl {{.ActiveSpec}}`.

## Session Close

Before ending a session: commit all changes, run quality gates (tests, build), update bead status, and push to remote (if configured). Work is not complete until changes are committed and pushed.
