package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var instructCmd = &cobra.Command{
	Use:   "instruct",
	Short: "Emit agent instructions for a spec or bead",
	Long:  `Generates and emits structured instructions for agent consumption. (Not yet implemented)`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("instruct: not yet implemented (see ADR-0003)")
	},
}

var nextCmd = &cobra.Command{
	Use:   "next",
	Short: "Show the next bead to work on",
	Long:  `Queries Beads for the next actionable work item. (Not yet implemented)`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("next: not yet implemented (see ADR-0003)")
	},
}

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate spec and plan artifacts",
	Long:  `Validates specification and plan documents against MindSpec conventions. (Not yet implemented)`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("validate: not yet implemented (see ADR-0003)")
	},
}
