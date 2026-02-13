# MindSpec Happy Path: End-to-End Feature Development

## Overview

MindSpec enforces a three-phase gated lifecycle: **Spec → Plan → Implement**. Every phase transition requires explicit human approval. The agent can never skip ahead.

```
Idle ──→ Spec Mode ──human gate──→ Plan Mode ──human gate──→ Implementation Mode ──→ Idle
```

---

## Phase 0: Idle

**State**: `mode: idle` (or no `.mindspec/state.json`)

On session start, the SessionStart hook runs `mindspec instruct`, which emits idle-mode guidance listing available specs and suggesting `/spec-init`.

---

## Phase 1: Spec Mode

### Human says
"I want to build feature X" or invokes `/spec-init`

### Agent does
1. Asks for spec ID (e.g. `009-feature-name`), title, and context
2. Creates `docs/specs/009-feature-name/spec.md` from template
3. Creates placeholder `context-pack.md`
4. Runs `mindspec state set --mode=spec --spec=009-feature-name`
5. Tells the human they're in Spec Mode

### Iterative collaboration
Agent and human fill in the spec — Goal, Requirements, Acceptance Criteria, Impacted Domains, ADR Touchpoints, Open Questions. **Only markdown artifacts are permitted** — no code, no tests.

### Available CLI commands
| Command | Purpose |
|---------|---------|
| `mindspec instruct` | Re-emit spec mode guidance |
| `mindspec validate spec <id>` | Structural quality check |
| `mindspec context pack <id>` | Generate context pack |

---

## Phase 2: Spec Approval (Human Gate)

### Human says
"Looks good, approve it" or invokes `/spec-approve`

### Agent does
1. Summarizes the spec and asks: "Do you approve this spec for planning?"
2. On **yes** — runs `mindspec approve spec 009-feature-name`

### What the CLI does
1. `validate.ValidateSpec()` — checks required sections, acceptance criteria quality
2. Updates `## Approval` section in spec.md → `Status: APPROVED`
3. Resolves spec gate `[GATE spec-approve 009-feature-name]` (best-effort)
4. Sets state → `{mode: "plan", activeSpec: "009-feature-name"}`
5. **Instruct-tail**: emits Plan Mode guidance

### Agent immediately begins planning
The `/spec-approve` skill instructs the agent to start planning right away (no second confirmation needed for entering plan mode — the approval **is** the authorization).

---

## Phase 3: Plan Mode

### Agent does
1. Reviews domain docs and accepted ADRs for impacted domains
2. Reviews Context Map for neighbor contracts
3. Creates `docs/specs/009-feature-name/plan.md` with YAML frontmatter (`status: Draft`)
4. Decomposes the spec into `work_chunks` — each with id, title, scope, verify steps, and dependencies
5. Iteratively refines with the human

### What a plan looks like
```yaml
---
status: Draft
spec_id: 009-feature-name
work_chunks:
  - id: 1
    title: "Core data model"
    scope: "internal/pkg/model.go"
    verify: ["Unit tests pass", "Struct matches spec schema"]
    depends_on: []
  - id: 2
    title: "CLI command wiring"
    scope: "cmd/mindspec/feature.go"
    verify: ["--help output correct", "Integration test passes"]
    depends_on: [1]
---
```

---

## Phase 4: Plan Approval (Human Gate)

### Human says
"Plan looks good" or invokes `/plan-approve`

### Agent does
1. Summarizes the plan (beads, scope, deps) and asks: "Do you approve this plan for implementation?"
2. On **yes** — runs `mindspec approve plan 009-feature-name`

### What the CLI does
1. `validate.ValidatePlan()` — checks frontmatter, work_chunks, verification steps
2. Updates plan frontmatter → `status: Approved`, `approved_at`, `approved_by`
3. Resolves plan gate `[GATE plan-approve 009-feature-name]` (best-effort)
4. Sets state → stays `plan` mode (deliberately NOT implement — need a bead first)
5. **Instruct-tail**: emits guidance telling user to run `mindspec next`

### The agent then tells the human
> Run `mindspec next` to claim the first ready bead and enter Implementation Mode.

---

## Phase 5: Claiming Work

### Agent (or human) runs
`mindspec next`

### What the CLI does
1. **Clean tree check** — fails if uncommitted changes
2. **Query ready work** — searches for molecule children via `bd ready --parent`, falls back to `bd ready`
3. **Display & select** — shows available beads, picks first (or `--pick=N`)
4. **Claim** — `bd update <id> in_progress`
5. **Create worktree** — `bd worktree create worktree-<beadID> bead/<beadID>`
6. **Resolve mode** — maps bead type to MindSpec mode (`task` → `implement`)
7. **Set state** → `{mode: "implement", activeSpec: "009-feature-name", activeBead: "<beadID>"}`
8. **Instruct-tail**: emits Implementation Mode guidance with bead scope and obligations

---

## Phase 6: Implementation

### Agent does (within worktree)
1. Writes code **within the bead's declared scope**
2. Creates tests
3. Updates documentation (**doc-sync is mandatory** — "done" includes doc-sync)
4. Follows cited ADRs (divergence → stop + inform human)
5. Uses commit convention: `impl(<bead-id>): <summary>`
6. Runs verification steps from the plan

### Constraints
- **Scope discipline**: new work becomes new beads, not scope creep
- **Worktree isolation**: work happens in `worktree-<beadID>`, not main
- **ADR compliance**: divergence triggers a human gate

---

## Phase 7: Bead Completion

### Agent runs
`mindspec complete`

### What the CLI does
1. Reads `activeBead` from state
2. Finds matching worktree
3. **Clean tree check** — all changes must be committed
4. **Close bead** — `bd close <beadID>`
5. **Remove worktree** — `bd worktree remove`
6. **Advance state**:
   - If more ready beads → stays `implement`, sets next bead
   - If beads exist but blocked → transitions to `plan`
   - If all beads done → transitions to `idle`
7. **Instruct-tail**: emits guidance for the new state

---

## Phase 8: Loop or Finish

If `mindspec complete` found another ready bead → run `mindspec next` again, repeat Phase 5–7.

If all beads are done → state goes to `idle`, the feature is complete.

---

## Summary: Who Does What

| Step | Human | Agent | CLI Command |
|------|-------|-------|-------------|
| Start feature | "Build X" | Creates spec, sets state | `mindspec state set --mode=spec` |
| Write spec | Reviews, guides | Writes markdown | — |
| Approve spec | "Yes, approved" | Runs approval | `mindspec approve spec <id>` |
| Write plan | Reviews plan | Decomposes into chunks | — |
| Approve plan | "Yes, approved" | Runs approval | `mindspec approve plan <id>` |
| Claim work | — | Claims bead | `mindspec next` |
| Implement | Reviews code | Codes + doc-sync | `impl(bead): ...` commits |
| Complete bead | — | Closes bead | `mindspec complete` |
| Next bead | — | Claims next | `mindspec next` (loop) |
| All done | — | State → idle | (automatic) |

---

## Gaps Found in the Current Implementation

### 1. Bead creation is a manual, undocumented step (Critical)

This is the biggest gap. The `approve plan` command validates and approves the plan, but **it does not create the implementation beads**. The beads (molecule parent + impl tasks + gates) are created by `mindspec bead plan <id>` — but that command is marked `Deprecated` with the note "use /plan-approve workflow instead, which calls this automatically." **But `/plan-approve` does NOT call it.** There's a lie in the deprecation message.

Similarly, `mindspec bead spec <id>` creates the spec bead and spec gate, but nothing in the `/spec-init` or `/spec-approve` workflows calls it. The approve commands gracefully handle missing gates (warn and proceed), which masks the gap rather than fixing it.

**In practice**: The agent has been creating beads manually during planning, or running `mindspec bead spec/plan` ad-hoc. This works but isn't orchestrated — it relies on the agent "knowing" to do it.

**Fix**: Either `approve plan` should call `CreatePlanBeads()` internally, or the `/plan-approve` skill should explicitly call `mindspec bead plan <id>` before `mindspec approve plan <id>`.

### 2. Spec ID parsing in `next` is broken for standard titles

`ResolveMode()` extracts spec ID from bead titles by looking for a colon, but the actual title convention is `[IMPL 009-feature.1] Chunk title` (no colon). This means spec ID is empty for impl beads resolved via `next`. In practice this doesn't cause failures because `task` type beads default to `implement` mode regardless, and the spec ID is often already in state from the prior `approve plan` call. But it's fragile.

### 3. Worktree mismatch warning fires immediately after `next`

`CheckWorktree()` compares CWD to the expected worktree path. Since `mindspec next` creates the worktree but the CWD is still the main repo, the instruct-tail always warns "you're not in the right worktree" immediately after claiming work. The agent sees a false alarm.

### 4. No automatic context pack generation

The `/spec-init` skill creates a placeholder `context-pack.md`. There's no trigger to generate the real one — the agent must remember to run `mindspec context pack <id>` manually. Neither `/spec-approve` nor `/plan-approve` generate it.

### 5. Hardcoded `approved_by: user`

Both approval commands hardcode `approved_by: "user"`. No mechanism to capture the actual approver identity.

### 6. Error message says "stash" but project conventions forbid it

`complete.go` tells the user to "commit or stash before completing" — but `AGENTS.md` says "Never auto-stash."

### 7. Idle-mode instruct is informational, not directive

The `idle.md` template lists available actions but doesn't tell the agent to actually do anything. The other mode templates (spec, plan, implement) all have a `## Next Action` section that gives the agent an explicit directive — idle is missing this. The result is that on a fresh session with no active work, the agent passively waits instead of proactively greeting the user and suggesting next steps (`/spec-init`, resume a spec, `mindspec doctor`).

**Fix**: Add a `## Next Action` section to `internal/instruct/templates/idle.md` with an agent directive, matching the convention in other templates.

### 8. No milestone commit orchestration

The mode docs and AGENTS.md describe milestone commits (`spec(<bead-id>): ...`, `plan(<bead-id>): ...`, `impl(<bead-id>): ...`), but the CLI commands don't create them. The agent is expected to make these commits manually after running the approve/complete commands. This is fine in principle but means the convention is only enforced by agent training, not tooling.

---

## Overall Assessment

The **three-mode gating system is solid** — the state machine, mode templates, and instruct-tail convention work well together to keep the agent on rails. The `/spec-approve` → planning → `/plan-approve` → `next` → implement → `complete` flow is coherent.

The main weakness is **the bead creation gap** (Gap #1). The workflow assumes beads exist when `next` queries for ready work, but no step in the automated flow actually creates them. This is the single most important thing to fix for a clean happy path. Everything else is minor polish.
