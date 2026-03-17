package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "doug-plan",
	Short: "doug-plan orchestrates multi-agent planning workflows",
}

func Execute() error {
	return rootCmd.Execute()
}
