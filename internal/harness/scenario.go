package harness

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

// Scenario defines a behavioral test scenario for an agent session.
type Scenario struct {
	Name        string                                                     // scenario identifier (e.g. "single_bead")
	Description string                                                     // human-readable description
	Setup       func(sandbox *Sandbox) error                               // prepares sandbox state before agent runs
	Prompt      string                                                     // the prompt given to the agent
	Assertions  func(t *testing.T, sandbox *Sandbox, events []ActionEvent) // post-run assertions
	MaxTurns    int                                                        // turn budget (0 = unlimited)
	Model       string                                                     // model override (e.g. "haiku")
}

// AllScenarios returns all defined behavior scenarios.
func AllScenarios() []Scenario {
	return []Scenario{
		ScenarioSpecToIdle(),
		ScenarioSingleBead(),
		ScenarioMultiBeadDeps(),
		ScenarioAbandonSpec(),
		ScenarioInterruptForBug(),
		ScenarioResumeAfterCrash(),
		ScenarioSpecInit(),
		ScenarioSpecApprove(),
		ScenarioPlanApprove(),
		ScenarioImplApprove(),
		ScenarioSpecStatus(),
	}
}

// ScenarioSpecToIdle tests the full lifecycle: explore → spec → plan → implement → review → idle.
func ScenarioSpecToIdle() Scenario {
	return Scenario{
		Name:        "spec_to_idle",
		Description: "Full lifecycle from explore through idle",
		MaxTurns:    75,
		Model:       "haiku",
		Setup: func(sandbox *Sandbox) error {
			// Sandbox comes with a clean repo; agent starts from scratch
			return nil
		},
		Prompt: `IMPORTANT: Do NOT respond conversationally. Do NOT ask what I'd like to do. Execute immediately.

You are in a MindSpec project with no active work. Your task: add a simple "greeting" feature — a hello.go program that prints "Hello!". Take it from idea all the way through to a completed implementation using the mindspec workflow.`,
		Assertions: func(t *testing.T, sandbox *Sandbox, events []ActionEvent) {
			// Agent may use explore+promote or go straight to spec-init — both are valid paths.
			assertCommandRanEither(t, events, "mindspec",
				[]string{"spec-init"}, []string{"explore", "promote"})
			assertCommandRan(t, events, "mindspec", "next")
			assertCommandRan(t, events, "mindspec", "complete")

			// Approve commands ran during lifecycle
			assertCommandRan(t, events, "mindspec", "approve")

			// Git state after full lifecycle
			assertBranchIs(t, sandbox, "main")
			assertNoBranches(t, sandbox, "spec/")
			assertNoBranches(t, sandbox, "bead/")
			assertNoWorktrees(t, sandbox)

			// Agent used worktrees during implementation
			assertEventCWDContains(t, events, ".worktrees/")
		},
	}
}

// ScenarioSingleBead tests implementing a single pre-approved bead.
func ScenarioSingleBead() Scenario {
	return Scenario{
		Name:        "single_bead",
		Description: "Pre-approved plan, implement a single bead",
		MaxTurns:    15,
		Model:       "haiku",
		Setup: func(sandbox *Sandbox) error {
			// Create real beads: epic + child task
			epicID := sandbox.CreateBead("[001-greeting] Epic", "epic", "")
			beadID := sandbox.CreateBead("[001-greeting] Implement greeting", "task", epicID)
			sandbox.ClaimBead(beadID)

			// Set up as if spec and plan are already approved
			sandbox.WriteFile(".mindspec/docs/specs/001-greeting/spec.md", `---
title: Greeting Feature
status: Approved
---
# Greeting Feature
Add a greeting function.
`)
			sandbox.WriteFile(".mindspec/docs/specs/001-greeting/plan.md", `---
status: Approved
spec_id: 001-greeting
---
# Plan
## Bead 1: Implement greeting
Create greeting.go with a Greet function.
`)
			sandbox.WriteFile(".mindspec/docs/specs/001-greeting/lifecycle.yaml",
				fmt.Sprintf("phase: implement\nepic_id: %s\n", epicID))
			sandbox.WriteFocus(mustJSON(map[string]string{
				"mode":       "implement",
				"activeSpec": "001-greeting",
				"activeBead": beadID,
				"timestamp":  time.Now().UTC().Format(time.RFC3339),
			}))
			sandbox.Commit("setup: pre-approved spec and plan")
			return nil
		},
		Prompt: `You are in implement mode for a pre-approved spec. A bead is already claimed.
Your task: create a file called greeting.go with a function Greet(name string) string
that returns "Hello, <name>!". Then run 'mindspec complete' to finish the bead.`,
		Assertions: func(t *testing.T, sandbox *Sandbox, events []ActionEvent) {
			// Agent should have created the file
			if !sandbox.FileExists("greeting.go") {
				t.Error("greeting.go was not created")
			}
			// Agent should have run mindspec complete
			assertCommandRan(t, events, "mindspec", "complete")
		},
	}
}

// ScenarioMultiBeadDeps tests implementing 3 beads with dependencies.
func ScenarioMultiBeadDeps() Scenario {
	return Scenario{
		Name:        "multi_bead_deps",
		Description: "Three beads with dependency chain",
		MaxTurns:    30,
		Model:       "haiku",
		Setup: func(sandbox *Sandbox) error {
			// Create real beads: epic + 3 child tasks
			epicID := sandbox.CreateBead("[002-multi] Epic", "epic", "")
			bead1 := sandbox.CreateBead("[002-multi] Core types", "task", epicID)
			sandbox.CreateBead("[002-multi] Formatter", "task", epicID)
			sandbox.CreateBead("[002-multi] Tests", "task", epicID)
			sandbox.ClaimBead(bead1)

			sandbox.WriteFile(".mindspec/docs/specs/002-multi/spec.md", `---
title: Multi-bead Feature
status: Approved
---
# Multi-bead Feature
Implement a feature in three phases.
`)
			sandbox.WriteFile(".mindspec/docs/specs/002-multi/plan.md", `---
status: Approved
spec_id: 002-multi
---
# Plan
## Bead 1: Core types
Create types.go with a Message struct.
## Bead 2: Formatter (depends on Bead 1)
Create formatter.go that formats Messages.
## Bead 3: Tests (depends on Bead 2)
Create formatter_test.go with tests.
`)
			sandbox.WriteFile(".mindspec/docs/specs/002-multi/lifecycle.yaml",
				fmt.Sprintf("phase: implement\nepic_id: %s\n", epicID))
			sandbox.WriteFocus(mustJSON(map[string]string{
				"mode":       "implement",
				"activeSpec": "002-multi",
				"activeBead": bead1,
				"timestamp":  time.Now().UTC().Format(time.RFC3339),
			}))
			sandbox.Commit("setup: multi-bead spec")
			return nil
		},
		Prompt: `You are in implement mode for a multi-bead spec. Implement all three beads in order:
1. Create types.go with a Message struct (fields: From, To, Body string)
2. Create formatter.go with FormatMessage(m Message) string
3. Create formatter_test.go that tests FormatMessage
Run 'mindspec complete' after each bead.`,
		Assertions: func(t *testing.T, sandbox *Sandbox, events []ActionEvent) {
			for _, f := range []string{"types.go", "formatter.go", "formatter_test.go"} {
				if !sandbox.FileExists(f) {
					t.Errorf("%s was not created", f)
				}
			}
		},
	}
}

// ScenarioAbandonSpec tests explore → dismiss flow.
func ScenarioAbandonSpec() Scenario {
	return Scenario{
		Name:        "abandon_spec",
		Description: "Enter explore mode and dismiss without promoting",
		MaxTurns:    10,
		Model:       "haiku",
		Setup: func(sandbox *Sandbox) error {
			return nil
		},
		Prompt: `IMPORTANT: Do NOT respond conversationally. Do NOT ask what I'd like to do. Execute immediately.

You are in a MindSpec project. Evaluate whether adding a "bad idea" feature is worth pursuing. After evaluating, decide it is not worth it and exit the exploration.`,
		Assertions: func(t *testing.T, sandbox *Sandbox, events []ActionEvent) {
			assertCommandRan(t, events, "mindspec", "explore")
			// Check that dismiss was called
			assertCommandContains(t, events, "mindspec", "dismiss")
		},
	}
}

// ScenarioInterruptForBug tests mid-bead interrupt for a bug fix.
func ScenarioInterruptForBug() Scenario {
	return Scenario{
		Name:        "interrupt_for_bug",
		Description: "Interrupt mid-bead to fix a bug, then resume",
		MaxTurns:    25,
		Model:       "haiku",
		Setup: func(sandbox *Sandbox) error {
			epicID := sandbox.CreateBead("[003-feature] Epic", "epic", "")
			beadID := sandbox.CreateBead("[003-feature] Implement feature", "task", epicID)
			sandbox.ClaimBead(beadID)

			sandbox.WriteFile(".mindspec/docs/specs/003-feature/lifecycle.yaml",
				fmt.Sprintf("phase: implement\nepic_id: %s\n", epicID))
			sandbox.WriteFocus(mustJSON(map[string]string{
				"mode":       "implement",
				"activeSpec": "003-feature",
				"activeBead": beadID,
				"timestamp":  time.Now().UTC().Format(time.RFC3339),
			}))
			sandbox.WriteFile("main.go", `package main

func main() {
	// existing code
}
`)
			sandbox.Commit("setup: feature in progress")
			return nil
		},
		Prompt: `You are implementing a feature bead. While working, you notice
main.go has a critical bug — the main function doesn't print anything.
Fix main.go to add fmt.Println("hello") and commit the fix, then continue your feature work
by creating feature.go with a Feature() function. Run 'mindspec complete' when done.`,
		Assertions: func(t *testing.T, sandbox *Sandbox, events []ActionEvent) {
			if !sandbox.FileExists("feature.go") {
				t.Error("feature.go was not created")
			}
		},
	}
}

// ScenarioResumeAfterCrash tests picking up an existing in-progress bead.
func ScenarioResumeAfterCrash() Scenario {
	return Scenario{
		Name:        "resume_after_crash",
		Description: "Resume an in-progress bead after session crash",
		MaxTurns:    15,
		Model:       "haiku",
		Setup: func(sandbox *Sandbox) error {
			epicID := sandbox.CreateBead("[004-resume] Epic", "epic", "")
			beadID := sandbox.CreateBead("[004-resume] Process feature", "task", epicID)
			sandbox.ClaimBead(beadID)

			// Simulate a crash: focus says implement, bead is in_progress, partial work exists
			sandbox.WriteFile(".mindspec/docs/specs/004-resume/lifecycle.yaml",
				fmt.Sprintf("phase: implement\nepic_id: %s\n", epicID))
			sandbox.WriteFocus(mustJSON(map[string]string{
				"mode":       "implement",
				"activeSpec": "004-resume",
				"activeBead": beadID,
				"timestamp":  time.Now().UTC().Format(time.RFC3339),
			}))
			sandbox.WriteFile("partial.go", `package main

// TODO: finish this function
func Process() {
}
`)
			sandbox.Commit("setup: partial work before crash")
			return nil
		},
		Prompt: `You are resuming after a session crash. The project is in implement mode with
a bead in progress. There's a partial.go file with an incomplete Process function.
Complete the Process function (make it return "processed") and run 'mindspec complete'.`,
		Assertions: func(t *testing.T, sandbox *Sandbox, events []ActionEvent) {
			if !sandbox.FileExists("partial.go") {
				t.Error("partial.go should still exist")
			}
			assertCommandRan(t, events, "mindspec", "complete")
		},
	}
}

// ScenarioSpecInit tests the /ms-spec-init flow: idle → spec-init → spec mode with worktree.
func ScenarioSpecInit() Scenario {
	return Scenario{
		Name:        "spec_init",
		Description: "Initialize a new spec from idle mode",
		MaxTurns:    15,
		Model:       "haiku",
		Setup: func(sandbox *Sandbox) error {
			// Sandbox starts in idle mode — no active spec
			return nil
		},
		Prompt: `IMPORTANT: Do NOT respond conversationally. Execute immediately.

/ms-spec-init 001-calculator --title "Calculator"`,
		Assertions: func(t *testing.T, sandbox *Sandbox, events []ActionEvent) {
			// Agent should have run mindspec spec-init
			assertCommandRan(t, events, "mindspec", "spec-init")

			// A spec branch should exist
			if branches := sandbox.ListBranches("spec/"); len(branches) == 0 {
				t.Error("expected a spec/ branch to be created")
			}

			// A worktree should exist
			if wts := sandbox.ListWorktrees(); len(wts) == 0 {
				t.Error("expected a worktree to be created")
			}
		},
	}
}

// ScenarioSpecApprove tests the /ms-spec-approve flow: spec mode → approve → plan mode.
func ScenarioSpecApprove() Scenario {
	return Scenario{
		Name:        "spec_approve",
		Description: "Approve a draft spec and transition to plan mode",
		MaxTurns:    15,
		Model:       "haiku",
		Setup: func(sandbox *Sandbox) error {
			specID := "001-calc"
			specBranch := "spec/" + specID

			// Create the branch and switch back to main
			mustRunSandbox(sandbox, "git", "branch", specBranch)

			// Create worktree from the branch
			wtDir := ".worktrees/worktree-spec-" + specID
			mustRunSandbox(sandbox, "git", "worktree", "add", wtDir, specBranch)

			// Create epic for lifecycle
			epicID := sandbox.CreateBead("["+specID+"] Epic", "epic", "")

			// Write spec file in the worktree
			sandbox.WriteFile(wtDir+"/.mindspec/docs/specs/"+specID+"/spec.md", `---
title: Calculator Feature
status: Draft
---
# Calculator Feature

## Summary
Add basic arithmetic operations.

## Motivation
Users need a calculator.

## Detailed Design
Implement add and subtract functions.

## Acceptance Criteria
- add(a, b) returns a + b
- subtract(a, b) returns a - b
`)
			// Write lifecycle in worktree
			sandbox.WriteFile(wtDir+"/.mindspec/docs/specs/"+specID+"/lifecycle.yaml",
				fmt.Sprintf("phase: spec\nepic_id: %s\n", epicID))

			// Commit in the worktree
			mustRunSandbox(sandbox, "git", "-C", wtDir, "add", "-A")
			mustRunSandbox(sandbox, "git", "-C", wtDir, "commit", "-m", "setup: draft spec")

			// Set focus to spec mode (in main repo)
			sandbox.WriteFocus(mustJSON(map[string]string{
				"mode":           "spec",
				"activeSpec":     specID,
				"specBranch":     specBranch,
				"activeWorktree": wtDir,
				"timestamp":      time.Now().UTC().Format(time.RFC3339),
			}))
			sandbox.Commit("setup: spec mode focus")
			return nil
		},
		Prompt: `IMPORTANT: Do NOT respond conversationally. Execute immediately.

/ms-spec-approve`,
		Assertions: func(t *testing.T, sandbox *Sandbox, events []ActionEvent) {
			// Agent should have run mindspec approve spec
			assertCommandRan(t, events, "mindspec", "approve")
			assertCommandContains(t, events, "mindspec", "spec")
		},
	}
}

// ScenarioPlanApprove tests the /ms-plan-approve flow: plan mode → approve → implement mode.
func ScenarioPlanApprove() Scenario {
	return Scenario{
		Name:        "plan_approve",
		Description: "Approve a plan and transition to implement mode",
		MaxTurns:    20,
		Model:       "haiku",
		Setup: func(sandbox *Sandbox) error {
			specID := "001-planner"
			specBranch := "spec/" + specID

			// Create epic
			epicID := sandbox.CreateBead("["+specID+"] Epic", "epic", "")

			// Create spec branch and worktree (stay on main)
			mustRunSandbox(sandbox, "git", "branch", specBranch)
			wtDir := ".worktrees/worktree-spec-" + specID
			mustRunSandbox(sandbox, "git", "worktree", "add", wtDir, specBranch)

			// Write approved spec
			sandbox.WriteFile(wtDir+"/.mindspec/docs/specs/"+specID+"/spec.md", `---
title: Planner Feature
status: Approved
---
# Planner Feature

## Summary
Add a planning feature.

## Acceptance Criteria
- plan() returns a plan string
`)
			// Write draft plan with bead sections
			sandbox.WriteFile(wtDir+"/.mindspec/docs/specs/"+specID+"/plan.md", `---
status: Draft
spec_id: 001-planner
---
# Plan

## Bead 1: Core planner
Create planner.go with a Plan() function.

## Bead 2: Tests
Create planner_test.go with tests.
Depends on: Bead 1
`)
			// Write lifecycle in plan phase
			sandbox.WriteFile(wtDir+"/.mindspec/docs/specs/"+specID+"/lifecycle.yaml",
				fmt.Sprintf("phase: plan\nepic_id: %s\n", epicID))

			// Set focus to plan mode
			sandbox.WriteFocus(mustJSON(map[string]string{
				"mode":           "plan",
				"activeSpec":     specID,
				"specBranch":     specBranch,
				"activeWorktree": wtDir,
				"timestamp":      time.Now().UTC().Format(time.RFC3339),
			}))

			// Commit in worktree
			mustRunSandbox(sandbox, "git", "-C", wtDir, "add", "-A")
			mustRunSandbox(sandbox, "git", "-C", wtDir, "commit", "-m", "setup: draft plan")

			// Commit focus in main
			sandbox.Commit("setup: plan mode focus")
			return nil
		},
		Prompt: `IMPORTANT: Do NOT respond conversationally. Execute immediately.

/ms-plan-approve`,
		Assertions: func(t *testing.T, sandbox *Sandbox, events []ActionEvent) {
			// Agent should have run mindspec approve plan
			assertCommandRan(t, events, "mindspec", "approve")

			// Agent should have run mindspec next (per /ms-plan-approve flow)
			assertCommandRan(t, events, "mindspec", "next")
		},
	}
}

// ScenarioImplApprove tests the /ms-impl-approve flow: review mode → approve impl → idle.
func ScenarioImplApprove() Scenario {
	return Scenario{
		Name:        "impl_approve",
		Description: "Approve implementation and transition to idle",
		MaxTurns:    15,
		Model:       "haiku",
		Setup: func(sandbox *Sandbox) error {
			specID := "001-done"
			specBranch := "spec/" + specID

			// Create epic + bead (already closed)
			epicID := sandbox.CreateBead("["+specID+"] Epic", "epic", "")
			beadID := sandbox.CreateBead("["+specID+"] Implement feature", "task", epicID)
			sandbox.ClaimBead(beadID)
			// Close the bead
			sandbox.runBDMust("close", beadID)

			// Create spec branch with implementation content
			mustRunSandbox(sandbox, "git", "checkout", "-b", specBranch)

			// Write spec files
			sandbox.WriteFile(".mindspec/docs/specs/"+specID+"/spec.md", `---
title: Done Feature
status: Approved
---
# Done Feature
A completed feature.
`)
			sandbox.WriteFile(".mindspec/docs/specs/"+specID+"/plan.md", fmt.Sprintf(`---
status: Approved
spec_id: %s
bead_ids:
- %s
---
# Plan
## Bead 1: Implement feature
Create done.go.
`, specID, beadID))
			sandbox.WriteFile(".mindspec/docs/specs/"+specID+"/lifecycle.yaml",
				fmt.Sprintf("phase: review\nepic_id: %s\n", epicID))

			// Write actual implementation file
			sandbox.WriteFile("done.go", `package main

func Done() string { return "done" }
`)
			sandbox.Commit("impl: implement feature")

			// Switch back to main for the merge target
			mustRunSandbox(sandbox, "git", "checkout", "main")

			// Set focus to review mode
			sandbox.WriteFocus(mustJSON(map[string]string{
				"mode":       "review",
				"activeSpec": specID,
				"specBranch": specBranch,
				"timestamp":  time.Now().UTC().Format(time.RFC3339),
			}))
			sandbox.Commit("setup: review mode focus")
			return nil
		},
		Prompt: `IMPORTANT: Do NOT respond conversationally. Execute immediately.

/ms-impl-approve`,
		Assertions: func(t *testing.T, sandbox *Sandbox, events []ActionEvent) {
			// Agent should have run mindspec approve impl
			assertCommandRan(t, events, "mindspec", "approve")
			assertCommandContains(t, events, "mindspec", "impl")
		},
	}
}

// ScenarioSpecStatus tests the /ms-spec-status flow: check current mode and report.
func ScenarioSpecStatus() Scenario {
	return Scenario{
		Name:        "spec_status",
		Description: "Check current MindSpec status and report mode",
		MaxTurns:    10,
		Model:       "haiku",
		Setup: func(sandbox *Sandbox) error {
			// Set up in implement mode so there's interesting state to report
			epicID := sandbox.CreateBead("[001-status] Epic", "epic", "")
			beadID := sandbox.CreateBead("[001-status] Feature", "task", epicID)
			sandbox.ClaimBead(beadID)

			sandbox.WriteFile(".mindspec/docs/specs/001-status/spec.md", `---
title: Status Feature
status: Approved
---
# Status Feature
`)
			sandbox.WriteFile(".mindspec/docs/specs/001-status/lifecycle.yaml",
				fmt.Sprintf("phase: implement\nepic_id: %s\n", epicID))
			sandbox.WriteFocus(mustJSON(map[string]string{
				"mode":       "implement",
				"activeSpec": "001-status",
				"activeBead": beadID,
				"timestamp":  time.Now().UTC().Format(time.RFC3339),
			}))
			sandbox.Commit("setup: implement mode with active bead")
			return nil
		},
		Prompt: `IMPORTANT: Do NOT respond conversationally. Execute immediately.

/ms-spec-status`,
		Assertions: func(t *testing.T, sandbox *Sandbox, events []ActionEvent) {
			// Agent should have run mindspec state show or mindspec instruct
			ran := false
			for _, e := range events {
				if e.Command != "mindspec" {
					continue
				}
				args := eventArgs(e)
				for _, arg := range args {
					if arg == "state" || arg == "instruct" {
						ran = true
						break
					}
				}
				if ran {
					break
				}
			}
			if !ran {
				t.Error("expected agent to run 'mindspec state show' or 'mindspec instruct'")
			}
		},
	}
}

// mustRunSandbox runs a command in the sandbox root, fataling on error.
func mustRunSandbox(sandbox *Sandbox, name string, args ...string) {
	sandbox.t.Helper()
	mustRun(sandbox.t, sandbox.Root, name, args...)
}

// --- Helpers ---

func mustJSON(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("mustJSON: %v", err))
	}
	return string(data)
}

func assertCommandRan(t *testing.T, events []ActionEvent, command string, argSubstr ...string) { //nolint:unparam // command kept for call-site clarity
	t.Helper()
	for _, e := range events {
		if e.Command != command {
			continue
		}
		if len(argSubstr) == 0 {
			return // found the command
		}
		args := eventArgs(e)
		if containsAll(args, argSubstr[0]) {
			return
		}
	}
	if len(argSubstr) > 0 {
		t.Errorf("command %q with arg %q was not found in events", command, argSubstr[0])
	} else {
		t.Errorf("command %q was not found in events", command)
	}
}

// assertCommandRanEither checks that the command was invoked with one of the
// given arg patterns (each is a list of substrings that must all appear).
func assertCommandRanEither(t *testing.T, events []ActionEvent, command string, patterns ...[]string) {
	t.Helper()
	for _, e := range events {
		if e.Command != command {
			continue
		}
		args := eventArgs(e)
		for _, pattern := range patterns {
			matched := true
			for _, sub := range pattern {
				if !containsAll(args, sub) {
					matched = false
					break
				}
			}
			if matched {
				return
			}
		}
	}
	t.Errorf("command %q was not found with any of the expected arg patterns %v", command, patterns)
}

func assertCommandContains(t *testing.T, events []ActionEvent, command, substr string) {
	t.Helper()
	for _, e := range events {
		if e.Command != command {
			continue
		}
		args := eventArgs(e)
		for _, arg := range args {
			if arg == substr {
				return
			}
		}
	}
	t.Errorf("command %q with arg containing %q was not found in events", command, substr)
}

// eventArgs returns args from both the Args map and ArgsList slice.
func eventArgs(e ActionEvent) []string {
	args := flatArgs(e.Args)
	args = append(args, e.ArgsList...)
	return args
}

func assertBranchIs(t *testing.T, sandbox *Sandbox, expected string) {
	t.Helper()
	actual := sandbox.GitBranch()
	if actual != expected {
		t.Errorf("expected current branch %q, got %q", expected, actual)
	}
}

func assertNoBranches(t *testing.T, sandbox *Sandbox, prefix string) {
	t.Helper()
	branches := sandbox.ListBranches(prefix)
	if len(branches) > 0 {
		t.Errorf("expected no branches with prefix %q, found: %v", prefix, branches)
	}
}

func assertNoWorktrees(t *testing.T, sandbox *Sandbox) {
	t.Helper()
	wts := sandbox.ListWorktrees()
	if len(wts) > 0 {
		t.Errorf("expected no worktrees, found: %v", wts)
	}
}

func assertEventCWDContains(t *testing.T, events []ActionEvent, substr string) {
	t.Helper()
	for _, e := range events {
		if strings.Contains(e.CWD, substr) {
			return
		}
	}
	t.Errorf("no event had CWD containing %q", substr)
}
