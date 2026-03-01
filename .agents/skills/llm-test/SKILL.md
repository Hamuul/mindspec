---
name: llm-test
description: Run the MindSpec LLM test harness command workflow defined in .claude/commands/llm-test.md. Use when the user asks to run llm tests, investigate harness failures, improve test scenarios, or execute the iterative test-observe-fix-retest loop.
---

# LLM Test

1. Read `.claude/commands/llm-test.md` from the current repository root.
2. Treat that file as the source of truth for workflow steps, required commands, and reporting.
3. Execute the requested llm-test task (single scenario, full suite, failure investigation, or scenario improvement) exactly as specified there.
4. Apply mandatory guardrails from the command file, including build prerequisites and required environment flags.
5. Summarize outcomes to the user with concrete run results and any follow-up changes.

If `.claude/commands/llm-test.md` is missing or renamed, stop and report the missing path before continuing.
