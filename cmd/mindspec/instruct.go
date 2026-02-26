package main

import (
	"fmt"
	"os"

	"github.com/mindspec/mindspec/internal/guard"
	"github.com/mindspec/mindspec/internal/instruct"
	"github.com/mindspec/mindspec/internal/resolve"
	"github.com/mindspec/mindspec/internal/state"
	"github.com/mindspec/mindspec/internal/trace"
	"github.com/mindspec/mindspec/internal/workspace"
	"github.com/spf13/cobra"
)

var instructCmd = &cobra.Command{
	Use:   "instruct",
	Short: "Emit agent instructions for the current mode and active work",
	Long: `Derives mode from the target spec's molecule state (ADR-0015) and emits
mode-appropriate operating guidance for agent consumption.

If --spec is omitted and exactly one active spec exists, it is auto-selected.
If multiple active specs exist, the command fails with a list of candidates.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		format, _ := cmd.Flags().GetString("format")
		specFlag, _ := cmd.Flags().GetString("spec")

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting working directory: %w", err)
		}

		root, err := workspace.FindRoot(cwd)
		if err != nil {
			return err
		}

		// CWD redirect: if running from main with an active worktree,
		// emit ONLY the redirect message — no normal guidance.
		if wtPath := guard.ActiveWorktreePath(root); wtPath != "" && guard.IsMainCWD(root) {
			msg := fmt.Sprintf("# MindSpec — CWD Redirect\n\nYou are in the main worktree. Run:\n\n  cd %s\n\nThen run `mindspec instruct` for mode-appropriate guidance.\n", wtPath)
			if format == "json" {
				fmt.Printf(`{"redirect":true,"worktree_path":%q,"message":"Switch to worktree"}`, wtPath)
				fmt.Println()
			} else {
				fmt.Print(msg)
			}
			return nil
		}

		// Resolve target spec (ADR-0015 targeting rules)
		specID, resolveErr := resolve.ResolveTarget(root, specFlag)

		// Build state from resolver or fall back to state.json
		var s *state.State
		if resolveErr != nil {
			// If ambiguous, surface the error directly
			if _, ok := resolveErr.(*resolve.ErrAmbiguousTarget); ok {
				return resolveErr
			}
			// Other errors: fall back to state.json
			s, err = state.Read(root)
			if err != nil {
				if err == state.ErrNoState {
					return handleNoState(root, format)
				}
				return err
			}
		} else {
			// Derive mode from molecule
			mode, modeErr := resolve.ResolveMode(root, specID)
			if modeErr != nil {
				// Fallback: read mode from state.json but use resolved specID
				s, _ = state.Read(root)
				if s == nil {
					s = &state.State{Mode: state.ModeIdle}
				}
				s.ActiveSpec = specID
			} else {
				// Read existing state for bead info, overlay derived values
				s, _ = state.Read(root)
				if s == nil {
					s = &state.State{}
				}
				s.Mode = mode
				s.ActiveSpec = specID
			}
		}

		ctx := instruct.BuildContext(root, s)

		// Add worktree check when an active worktree is set.
		if s.ActiveWorktree != "" {
			if warning := instruct.CheckWorktree(s.ActiveWorktree); warning != "" {
				ctx.Warnings = append(ctx.Warnings, "[worktree] "+warning)
			}
		}

		if format == "json" {
			output, err := instruct.RenderJSON(ctx)
			if err != nil {
				return err
			}
			fmt.Println(output)
			return nil
		}

		output, err := instruct.Render(ctx)
		if err != nil {
			return err
		}
		trace.Emit(trace.NewEvent("instruct.render").
			WithSpec(s.ActiveSpec).
			WithTokens(trace.EstimateTokens(output)).
			WithData(map[string]any{
				"tokens_total": trace.EstimateTokens(output),
				"mode":         s.Mode,
				"template":     s.Mode + ".md",
			}))
		fmt.Print(output)
		return nil
	},
}

// handleNoState provides a graceful fallback when state.json is missing.
func handleNoState(root, format string) error {
	// Fall back to idle mode with a warning
	s := &state.State{Mode: state.ModeIdle}
	ctx := instruct.BuildContext(root, s)
	ctx.Warnings = append(ctx.Warnings,
		"[state] No .mindspec/state.json found. Run `mindspec state set --mode=<mode> --spec=<id>` to initialize.",
	)

	if format == "json" {
		output, err := instruct.RenderJSON(ctx)
		if err != nil {
			return err
		}
		fmt.Println(output)
		return nil
	}

	output, err := instruct.Render(ctx)
	if err != nil {
		return err
	}
	fmt.Print(output)
	return nil
}

func init() {
	instructCmd.Flags().String("format", "", "Output format: markdown (default) or json")
	instructCmd.Flags().String("spec", "", "Target spec ID (auto-detected if exactly one active spec)")
}
