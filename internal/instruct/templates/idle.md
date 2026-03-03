# MindSpec — No Active Work

You are not currently working on any spec or bead.

## MindSpec Lifecycle

```
>>> idle ── spec ── plan ── implement ── review ── idle
```

| Phase | Command | What happens |
|-------|---------|--------------|
| idle → spec | `mindspec spec create <slug>` | Creates branch + worktree + spec template |
| spec → plan | `mindspec spec approve <id>` | Validates spec, auto-commits |
| plan → impl | `mindspec plan approve <id>` | Validates plan, auto-creates beads, auto-commits |
| per bead | `mindspec next` | Claims bead, creates bead worktree |
| bead done | `mindspec complete "msg"` | Auto-commits, closes bead, merges bead→spec, removes worktree |
| review → idle | `mindspec impl approve <id>` | Merges spec→main, removes all worktrees + branches |

### Git rules
- Do NOT run any raw git commands — all git operations are handled by mindspec
- This includes: `git commit`, `git merge`, `git pull`, `git rebase`, `git branch`, `git checkout`
{{- if .BranchProtection}}

## Branch Protection

**main is protected.** You MUST NOT edit files or commit while on main. All changes require a branch.

All new features and non-trivial changes MUST go through `mindspec spec create`. The spec-create command creates the branch + worktree automatically.
{{- end}}

## Available Actions

- `mindspec explore "idea"` — evaluate whether an idea is worth pursuing
- `mindspec spec create <slug>` — start a new specification (creates branch + worktree)
- `mindspec state set --mode=spec --spec=<id>` — resume work on an existing spec
- `mindspec doctor` — check project health

### Evaluating Ideas

Use `mindspec explore "idea"` to start a lightweight, conversational exploration. This does NOT change mode — it simply helps you evaluate whether an idea is worth pursuing.

When exploration reaches a conclusion:
- **Worth pursuing**: Run `mindspec explore promote <NNN-slug>` to create a spec
- **Not worth pursuing**: Run `mindspec explore dismiss` to exit. Use `--adr` flag to capture the decision as an ADR.

## Available Specs

{{- if .AvailableSpecs}}
{{range .AvailableSpecs}}
- `{{.}}`
{{- end}}
{{- else}}
No specs found in `.mindspec/docs/specs/`.
{{- end}}

## Next Action

If the user already gave a concrete task, execute it immediately.
- Do NOT greet or ask what they'd like to work on first.
- Do NOT claim success unless commands actually ran and exited 0.

If the user did NOT give a concrete task, do this in your first message:

1. Greet the user
2. Suggest these options directly:
   - `mindspec explore "idea"` to explore whether an idea is worth pursuing
   - `mindspec spec create <slug>` to draft a new specification (if they already know what to build)
   - Resuming an existing spec (if any are listed above)
   - `mindspec doctor` to check project health
3. Ask what they'd like to work on
