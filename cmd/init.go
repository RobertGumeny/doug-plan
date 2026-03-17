package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var agents string

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new doug-plan project",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("doug-plan init: not yet implemented")
		return nil
	},
}

func init() {
	initCmd.Flags().StringVar(&agents, "agents", "", "comma-separated list of agents")
	rootCmd.AddCommand(initCmd)
}
