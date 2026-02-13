---
description: Request approval to transition from Spec Mode to Plan Mode
---

# Spec Approval Workflow

Use this workflow to gate the transition from Spec Mode to Plan Mode.

## Trigger

User invokes `/spec-approve` or expresses readiness to plan.

## Steps

### 1. Identify Active Spec

Find the active spec. If unclear, ask the user which spec they want to approve.

### 2. Run Automated Validation

Run `mindspec validate spec <id>` to check structural quality. If there are errors:

> **Spec not ready for approval**
>
> <paste validate output>
>
> Please update `docs/specs/<id>/spec.md` and try again.

Remain in Spec Mode.

### 3. Read and Review Spec

Read `docs/specs/<id>/spec.md` and review the content for completeness (the automated checks cover structure; you review substance).

### 4. Present Spec Summary

If validation passes, present a summary:

> **Spec Summary: <id>**
>
> **Goal**: <goal summary>
>
> **Impacted Domains**: <domain list>
>
> **ADR Touchpoints**: <ADR list>
>
> **Scope**: <key files/components>
>
> **Acceptance Criteria** (<N> items):
> - <criterion 1>
> - <criterion 2>
> - ...
>
> **Ready to approve and begin planning?**

### 5. Request Explicit Approval

Ask the user:

> Do you approve this spec for planning? (yes/no)

### 6. On Approval

If user approves:

1. Update `docs/specs/<id>/spec.md` Approval section:
   ```markdown
   ## Approval

   - **Status**: APPROVED
   - **Approved By**: user
   - **Approval Date**: <today's date>
   - **Notes**: Approved via /spec-approve workflow
   ```

2. Update MindSpec state to Plan Mode (**before** the milestone commit):
   ```bash
   mindspec state set --mode=plan --spec=<id>
   ```

3. Inform user briefly:
   > **Spec approved!** Entering Plan Mode — beginning plan draft.

4. **Immediately begin planning** — do NOT ask "shall I proceed?" or wait for further confirmation. The spec approval IS the authorization to start planning. Proceed directly to:
   - Review domain docs and accepted ADRs for impacted domains
   - Check Context Map for neighboring context contracts
   - Decompose spec into implementation beads (bounded work chunks)
   - Define verification steps for each bead
   - Draft `docs/specs/<id>/plan.md`
   - When the plan draft is complete, inform the user and advise them to use `/plan-approve` when ready

### 7. On Rejection

If user declines:

> Spec remains in **DRAFT** status.
>
> What changes would you like to make before approval?

Remain in Spec Mode.

---

## Notes

- This is a human gate — the user must explicitly confirm spec approval
- Once approved, planning starts automatically (no second confirmation needed)
- Approval is recorded in git (the spec file itself)
- Re-approval is required if the spec is modified after approval
- This transitions to Plan Mode and immediately begins plan drafting
