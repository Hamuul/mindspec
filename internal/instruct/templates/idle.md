# MindSpec — No Active Work

You are not currently working on any spec or bead.

## Available Actions

- Run `/spec-init` to start a new specification
- Run `mindspec state set --mode=spec --spec=<id>` to resume work on an existing spec
- Run `mindspec doctor` to check project health

## Available Specs

{{- if .AvailableSpecs}}
{{range .AvailableSpecs}}
- `{{.}}`
{{- end}}
{{- else}}
No specs found in `docs/specs/`.
{{- end}}

## Next Action

Greet the user and ask what they'd like to work on. Suggest:
- `/spec-init` to draft a new specification
- Resuming an existing spec (if any are listed above)
- `mindspec doctor` to check project health
