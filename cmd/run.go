package cmd

import (
	"os"

	"github.com/robertgumeny/doug-plan/internal/orchestrator"
	"github.com/spf13/cobra"
)

var approvalFlag string

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the configured plan steps",
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		return orchestrator.Run(orchestrator.Options{
			ProjectRoot:  cwd,
			Out:          os.Stdout,
			In:           os.Stdin,
			ApprovalMode: approvalFlag,
		})
	},
}

func init() {
	runCmd.Flags().StringVar(&approvalFlag, "approval", "", "approval mode override: auto, soft, or hard")
	rootCmd.AddCommand(runCmd)
}
