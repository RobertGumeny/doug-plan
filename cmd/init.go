package cmd

import (
	"os"

	"github.com/robertgumeny/doug-plan/internal/scaffold"
	"github.com/spf13/cobra"
)

var agents string

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new doug-plan project",
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		return scaffold.Run(scaffold.Options{
			ProjectRoot: cwd,
			Agents:      scaffold.ParseAgents(agents),
			Out:         os.Stdout,
		})
	},
}

func init() {
	initCmd.Flags().StringVar(&agents, "agents", "", "comma-separated list of agents")
	rootCmd.AddCommand(initCmd)
}
