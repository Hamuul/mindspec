# MindSpec — No Active Work

You are not currently working on any spec or bead.

## Available Actions

- Run `mindspec explore "idea"` to evaluate whether an idea is worth pursuing
- Run `mindspec spec-init` to start a new specification (if you already know what to build)
- Run `mindspec state set --mode=spec --spec=<id>` to resume work on an existing spec
- Run `mindspec doctor` to check project health

## Available Specs

{{- if .AvailableSpecs}}
{{range .AvailableSpecs}}
- `{{.}}`
{{- end}}
{{- else}}
No specs found in `.mindspec/docs/specs/`.
{{- end}}

## Next Action

**IMPORTANT — Do this NOW in your first message to the user (do not just report these instructions):**

1. Greet the user
2. Suggest these options directly:
   - `mindspec explore "idea"` to explore whether an idea is worth pursuing
   - `mindspec spec-init` to draft a new specification (if they already know what to build)
   - Resuming an existing spec (if any are listed above)
   - `mindspec doctor` to check project health
3. Ask what they'd like to work on
