package main

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mindspec",
	Short: "MindSpec: Spec-Driven Development and Self-Documentation System",
	Long:  `MindSpec is a spec-driven development + context management framework.`,
}

func init() {
	rootCmd.AddCommand(doctorCmd)
	rootCmd.AddCommand(instructCmd)
	rootCmd.AddCommand(nextCmd)
	rootCmd.AddCommand(validateCmd)
}
