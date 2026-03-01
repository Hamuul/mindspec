package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mindspec/mindspec/internal/adr"
	"github.com/mindspec/mindspec/internal/explore"
	"github.com/mindspec/mindspec/internal/workspace"
	"github.com/spf13/cobra"
)

var exploreCmd = &cobra.Command{
	Use:   "explore [description]",
	Short: "Enter Explore Mode to evaluate whether an idea is worth pursuing",
	Long: `Enters Explore Mode — a lightweight, conversational pre-spec phase.
The LLM guides you through problem clarification, prior art discovery,
feasibility assessment, and a recommendation.

Use 'explore dismiss' to exit without a spec, or 'explore promote' to
create a spec from the exploration.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := findLocalRoot()
		if err != nil {
			return err
		}

		description := ""
		if len(args) > 0 {
			description = args[0]
		}

		if err := explore.Enter(root, description); err != nil {
			return err
		}

		if description != "" {
			fmt.Printf("Exploring: %s\n\n", description)
		} else {
			fmt.Println("Entered Explore Mode.")
		}

		if err := emitInstruct(root); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not emit guidance: %v\n", err)
		}

		return nil
	},
}

var exploreDismissCmd = &cobra.Command{
	Use:   "dismiss",
	Short: "Exit Explore Mode without creating a spec",
	Long: `Exits Explore Mode and returns to idle. Use --adr to capture the
decision as an Architecture Decision Record.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := findLocalRoot()
		if err != nil {
			return err
		}

		if err := explore.Dismiss(root); err != nil {
			return err
		}

		fmt.Println("Exploration dismissed. Returned to idle.")

		createADR, _ := cmd.Flags().GetBool("adr")
		if createADR {
			title, _ := cmd.Flags().GetString("title")
			domain, _ := cmd.Flags().GetString("domain")

			if title == "" {
				return fmt.Errorf("--title is required when using --adr")
			}

			var domains []string
			if domain != "" {
				for _, d := range strings.Split(domain, ",") {
					d = strings.TrimSpace(d)
					if d != "" {
						domains = append(domains, d)
					}
				}
			}

			path, err := adr.Create(root, title, adr.CreateOpts{
				Domains: domains,
			})
			if err != nil {
				return fmt.Errorf("creating ADR: %w", err)
			}

			relPath, relErr := filepath.Rel(root, path)
			if relErr != nil {
				relPath = path
			}
			fmt.Printf("ADR created: %s\n", filepath.ToSlash(relPath))
			fmt.Println("Fill in the Context and Decision sections to capture why this idea was dismissed.")
		}

		return nil
	},
}

var explorePromoteCmd = &cobra.Command{
	Use:   "promote <spec-id>",
	Short: "Promote exploration to a spec and enter Spec Mode",
	Long: `Creates a new spec from the exploration and enters Spec Mode.
The spec-id should follow the NNN-slug convention (e.g., 042-api-caching).`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := findLocalRoot()
		if err != nil {
			return err
		}

		specID := args[0]
		title, _ := cmd.Flags().GetString("title")

		if err := explore.Promote(root, specID, title); err != nil {
			return err
		}

		specPath := filepath.Join(workspace.SpecDir(root, specID), "spec.md")
		relPath, relErr := filepath.Rel(root, specPath)
		if relErr != nil {
			relPath = specPath
		}
		fmt.Printf("Exploration promoted to spec: %s\n\n", filepath.ToSlash(relPath))

		if err := emitInstruct(root); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not emit guidance: %v\n", err)
		}

		return nil
	},
}

func init() {
	exploreDismissCmd.Flags().Bool("adr", false, "Create an ADR capturing the dismissal decision")
	exploreDismissCmd.Flags().String("title", "", "ADR title (required with --adr)")
	exploreDismissCmd.Flags().String("domain", "", "ADR domain(s), comma-separated")

	explorePromoteCmd.Flags().String("title", "", "Spec title (derived from slug if omitted)")

	exploreCmd.AddCommand(exploreDismissCmd)
	exploreCmd.AddCommand(explorePromoteCmd)
}
