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
