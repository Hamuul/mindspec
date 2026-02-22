package templates

// DomainTemplateFileNames lists the domain template files used for domain scaffolding.
var DomainTemplateFileNames = []string{
	"overview.md",
	"architecture.md",
	"interfaces.md",
	"runbook.md",
}

// Spec returns the built-in spec template.
func Spec() string {
	return specTemplate
}

// Plan returns the built-in plan template.
func Plan() string {
	return planTemplate
}

// ADR returns the built-in ADR template.
func ADR() string {
	return adrTemplate
}

// Domain returns the built-in domain template by filename.
func Domain(filename string) string {
	switch filename {
	case "overview.md":
		return `# {{.DomainName}} Domain — Overview

## What This Domain Owns

<List of responsibilities and capabilities this domain owns.>

## Boundaries

<What this domain does NOT own.>

## Key Files

| File | Purpose |
|:-----|:--------|
| <path> | <purpose> |

## Current State

<Brief description of implementation status.>
`
	case "architecture.md":
		return `# {{.DomainName}} Domain — Architecture

## Key Patterns

<Describe the key architectural patterns and design choices in this domain.>

## Invariants

1. <Invariant 1 — a property that must always hold>
`
	case "interfaces.md":
		return `# {{.DomainName}} Domain — Interfaces

## Provided Interfaces

<APIs, contracts, or capabilities this domain exposes to other domains.>

## Consumed Interfaces

<Interfaces from other domains that this domain depends on.>

## Events

<Events this domain emits or subscribes to.>
`
	case "runbook.md":
		return `# {{.DomainName}} Domain — Runbook

## Common Operations

<Step-by-step instructions for common tasks in this domain.>

## Troubleshooting

<Common issues and how to resolve them.>
`
	default:
		return ""
	}
}

const specTemplate = `---
molecule_id: ""
status: Draft
approved_at: ""
approved_by: ""
---
# Spec <ID>: <Title>

## Goal

<Brief description of what this spec achieves and the target user outcome>

## Background

<Context, motivation, and any relevant prior decisions>

## Impacted Domains

- <domain-1>: <how it is impacted>

## ADR Touchpoints

- [ADR-NNNN](../../adr/ADR-NNNN.md): <why this ADR is relevant>

## Requirements

1. <Requirement 1>
2. <Requirement 2>

## Scope

### In Scope
- <File or component 1>

### Out of Scope
- <Explicitly excluded items>

## Non-Goals

- <What this spec intentionally does not address>

## Acceptance Criteria

- [ ] <Specific, measurable criterion 1>
- [ ] <Specific, measurable criterion 2>

## Validation Proofs

- <command 1>: <Expected outcome>

## Open Questions

- [ ] <Question that must be resolved before planning>

## Approval

- **Status**: DRAFT
- **Approved By**: -
- **Approval Date**: -
- **Notes**: -
`

const planTemplate = `---
status: Draft
spec_id: <NNN-slug>
version: "0.1"
last_updated: YYYY-MM-DD
work_chunks:
  - id: 1
    title: "<Short title for first chunk>"
    scope: "<Files or components this chunk delivers>"
    verify:
      - "<Specific, testable verification step>"
    depends_on: []
  - id: 2
    title: "<Short title for second chunk>"
    scope: "<Files or components>"
    verify:
      - "<Verification step>"
    depends_on: [1]
---

# Plan: Spec <NNN> — <Title>

**Spec**: [spec.md](spec.md)

---

## Bead <NNN>-A: <Short title>

**Scope**: <What this bead delivers>

**Steps**:
1. <Step 1>
2. <Step 2>
3. <Step 3>

**Verification**:
- [ ] <Specific, testable criterion>

**Depends on**: nothing

---

## Dependency Graph

<NNN>-A (<short description>)
  -> <NNN>-B (<short description>)
`

// SpecLifecycleFormula returns the spec-lifecycle formula TOML for Beads.
func SpecLifecycleFormula() string {
	return specLifecycleFormula
}

const adrTemplate = `# ADR-NNNN: <Title>

- **Date**: <YYYY-MM-DD>
- **Status**: Proposed
- **Domain(s)**: <comma-separated list>
- **Deciders**: <who decides>
- **Supersedes**: n/a
- **Superseded-by**: n/a

## Context

<What is the issue that we're seeing that motivates this decision or change?>

## Decision

<What is the change that we're proposing and/or doing?>

## Decision Details

<Detailed breakdown of the decision. Use subsections as needed.>

## Consequences

### Positive

- <Positive consequence 1>
- <Positive consequence 2>

### Negative / Tradeoffs

- <Negative consequence or tradeoff 1>
- <Negative consequence or tradeoff 2>

## Alternatives Considered

### 1. <Alternative name>

<Description and why it was rejected.>

## Validation / Rollout

1. <Validation step 1>
2. <Validation step 2>
`

const specLifecycleFormula = `# MindSpec Spec Lifecycle Formula
#
# Orchestrates the full spec-driven development lifecycle with human gates.
# Each phase produces a durable artifact; human gates block progression until
# the artifact is reviewed and approved.
#
# Phases:
#   1. Spec:      Write the specification (agent work)
#   2. Gate:      Approve spec (human review)
#   3. Plan:      Write the implementation plan (agent work)
#   4. Gate:      Approve plan (human review)
#   5. Implement: Build the feature in beads (agent work)
#   6. Gate:      Review implementation (human review)

formula = "spec-lifecycle"
description = """
MindSpec spec lifecycle: spec → approve → plan → approve → implement → review.

This formula tracks the full lifecycle of a single specification from drafting
through implementation and final review. Human gates ensure quality checkpoints
between phases.

## Usage

Create a molecule for a new spec:
` + "```bash" + `
bd mol wisp create spec-lifecycle --var spec_id=039-my-feature
` + "```" + `

Or for an existing spec that's already in progress:
` + "```bash" + `
bd mol wisp create spec-lifecycle --var spec_id=038-beads-native-multi-spec-state
` + "```" + `

## Artifacts Produced

Each phase creates or updates files under ` + "`.mindspec/docs/specs/{{spec_id}}/`" + `:
- ` + "`spec.md`" + ` — the specification (Phase 1)
- ` + "`context-pack.md`" + ` — generated context bundle (Phase 2, on approval)
- ` + "`plan.md`" + ` — the implementation plan (Phase 3)
- ` + "`proofs/`" + ` — optional proof outputs (Phase 5)
"""
type = "workflow"
version = 1

[vars.spec_id]
description = "The spec identifier, e.g. 039-my-feature (NNN-slug format)"
required = true

# =============================================================================
# Phase 1: Specification
# =============================================================================

[[steps]]
id = "spec"
title = "Write spec {{spec_id}}"
description = """
Draft the specification document.

**Scaffold (if new):**
` + "```bash" + `
mindspec spec-init {{spec_id}}
` + "```" + `

**Resume (if existing):**
` + "```bash" + `
mindspec state set --mode=spec --spec={{spec_id}}
` + "```" + `

Run ` + "`mindspec instruct --spec={{spec_id}}`" + ` for detailed writing guidance.

**Done when:** ` + "`mindspec validate spec {{spec_id}}`" + ` passes with no errors.
"""

# =============================================================================
# Gate 1: Spec Approval
# =============================================================================

[[steps]]
id = "spec-approve"
title = "Approve spec {{spec_id}}"
type = "human"
needs = ["spec"]
description = """
Human review gate: approve the specification.

` + "```bash" + `
mindspec approve spec {{spec_id}}
` + "```" + `

This validates the spec, updates the Approval section, creates/resolves the
spec gate in Beads, generates the context pack, and transitions state to
plan mode.
"""

# =============================================================================
# Phase 2: Plan
# =============================================================================

[[steps]]
id = "plan"
title = "Write plan {{spec_id}}"
needs = ["spec-approve"]
description = """
Draft the implementation plan.

State should already be in plan mode (set by ` + "`mindspec approve spec`" + `).
Create ` + "`plan.md`" + ` at ` + "`.mindspec/docs/specs/{{spec_id}}/plan.md`" + `.

Run ` + "`mindspec instruct --spec={{spec_id}}`" + ` for detailed planning guidance.

**Done when:** ` + "`mindspec validate plan {{spec_id}}`" + ` passes.
"""

# =============================================================================
# Gate 2: Plan Approval
# =============================================================================

[[steps]]
id = "plan-approve"
title = "Approve plan {{spec_id}}"
type = "human"
needs = ["plan"]
description = """
Human review gate: approve the implementation plan.

` + "```bash" + `
mindspec approve plan {{spec_id}}
` + "```" + `

This validates the plan, updates frontmatter to Approved, creates
implementation beads, writes bead IDs back into plan.md, and resolves the
plan gate.

**After approval:** run ` + "`mindspec next --spec={{spec_id}}`" + ` to claim the first
bead and enter implementation mode.
"""

# =============================================================================
# Phase 3: Implementation
# =============================================================================

[[steps]]
id = "implement"
title = "Implement {{spec_id}}"
needs = ["plan-approve"]
description = """
Implement the approved plan by working through beads.

**Bead cycle:**
1. ` + "`mindspec next --spec={{spec_id}}`" + ` — claim next ready bead
2. Write code, tests, docs in the worktree
3. ` + "`mindspec complete --spec={{spec_id}}`" + ` — close bead, advance state

Repeat until all beads are closed. ` + "`mindspec complete`" + ` transitions to review
mode when the last bead is done.

Run ` + "`mindspec instruct --spec={{spec_id}}`" + ` for implementation guidance.
"""

# =============================================================================
# Gate 3: Implementation Review
# =============================================================================

[[steps]]
id = "review"
title = "Review {{spec_id}}"
type = "human"
needs = ["implement"]
description = """
Human review gate: review the completed implementation.

` + "```bash" + `
mindspec approve impl {{spec_id}}
` + "```" + `

This verifies review mode is active and transitions state to idle.
The spec lifecycle is complete.
"""
`
