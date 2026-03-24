package cmd

import (
	"os"

	"github.com/robertgumeny/doug-plan/internal/orchestrator"
	"github.com/spf13/cobra"
)

var approvalFlag string
var rerunFlag string
var freshFlag bool

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
			RerunStage:   rerunFlag,
			Fresh:        freshFlag,
		})
	},
}

func init() {
	runCmd.Flags().StringVar(&approvalFlag, "approval", "", "approval mode override: auto, cli, or browser")
	runCmd.Flags().StringVar(&rerunFlag, "rerun", "", "re-run from stage: Discovery, Roadmapping, PRD, or Tasks")
	runCmd.Flags().BoolVar(&freshFlag, "fresh", false, "start fresh: clear all plan artifacts and begin at Discovery")
	rootCmd.AddCommand(runCmd)
}
