package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the configured plan steps",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("no steps configured")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
