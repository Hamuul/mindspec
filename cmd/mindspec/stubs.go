package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate spec and plan artifacts",
	Long:  `Validates specification and plan documents against MindSpec conventions. (Not yet implemented)`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("validate: not yet implemented (see ADR-0003)")
	},
}
