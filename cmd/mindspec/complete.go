package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/mindspec/mindspec/internal/bead"
	"github.com/mindspec/mindspec/internal/complete"
	"github.com/mindspec/mindspec/internal/guard"
	"github.com/mindspec/mindspec/internal/validate"
	"github.com/mindspec/mindspec/internal/workspace"
	"github.com/spf13/cobra"
)

var completeCmd = &cobra.Command{
	Use:   "complete [bead-id]",
	Short: "Close a bead, remove its worktree, and advance state",
	Long: `Orchestrates the full bead close-out:
  1. Auto-commits changes if a commit message is provided
  2. Validates all changes are committed (clean worktree)
  3. Closes the bead via bd close
  4. Removes the worktree via bd worktree remove
  5. Advances state (next bead, plan, or idle)

Usage:
  mindspec complete "describe what you did"    # auto-commit + close
  mindspec complete                            # close (tree must be clean)

The bead ID is auto-resolved from state if not provided.`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := findRoot()
		if err != nil {
			return err
		}
		specID, _ := cmd.Flags().GetString("spec")
		allowMain, _ := cmd.Flags().GetBool("allow-main")

		// Worktree scoping guard
		if !allowMain {
			cwd, _ := os.Getwd()
			kind, _, _ := workspace.DetectWorktreeContext(cwd)
			switch kind {
			case workspace.WorktreeMain:
				return fmt.Errorf("mindspec complete must run from a bead worktree.\nUse `mindspec next` to claim a bead and create a worktree first.\nUse --allow-main to bypass this check for recovery.")
			case workspace.WorktreeSpec:
				return fmt.Errorf("you're in a spec worktree — run `mindspec next` to claim a bead first, then `mindspec complete` from the bead worktree")
			}
		}

		// CWD guard: prefer running from the bead worktree.
		// If we're in main but an active worktree exists, auto-chdir there.
		if guard.IsMainCWD(root) {
			if wtPath := guard.ActiveWorktreePath(root); wtPath != "" {
				if err := os.Chdir(wtPath); err != nil {
					return fmt.Errorf("could not switch to active worktree %s: %w", wtPath, err)
				}
				fmt.Fprintf(os.Stderr, "note: switched to worktree %s\n", wtPath)
			}
		}

		if err := bead.Preflight(root); err != nil {
			fmt.Fprintf(os.Stderr, "preflight failed: %v\n", err)
			os.Exit(1)
		}

		// Parse args: first arg may be a bead ID or part of a commit message.
		// If it looks like a spec/bead ID, treat as bead ID; otherwise treat all args as commit message.
		var beadID, commitMsg string
		if len(args) > 0 {
			if validate.SpecID(args[0]) == nil || strings.HasPrefix(args[0], "mindspec-") {
				beadID = args[0]
				if len(args) > 1 {
					commitMsg = strings.Join(args[1:], " ")
				}
			} else {
				commitMsg = strings.Join(args, " ")
			}
		}

		result, err := complete.Run(root, beadID, specID, commitMsg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		fmt.Print(complete.FormatResult(result))

		// Instruct-tail: emit guidance for the new mode
		fmt.Println() // separator between summary and guidance
		if err := emitInstruct(root); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not emit guidance: %v\n", err)
		}
		return nil
	},
}

func init() {
	completeCmd.Flags().String("spec", "", "Target spec ID when multiple specs are active")
	completeCmd.Flags().Bool("allow-main", false, "Bypass the bead-worktree requirement (for recovery)")
}
