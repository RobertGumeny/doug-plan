package cmd

import (
	"io"
	"os"

	"github.com/robertgumeny/doug-plan/internal/prompt"
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
		selectedAgents, err := resolveAgents(agents, os.Stdout, os.Stdin, prompt.IsTTY(os.Stdin))
		if err != nil {
			return err
		}
		return scaffold.Run(scaffold.Options{
			ProjectRoot: cwd,
			Agents:      selectedAgents,
			Out:         os.Stdout,
		})
	},
}

// resolveAgents returns agents from the --agents flag if set, or prompts the
// user interactively on a TTY. When isTTY is false and no flag is provided it
// silently returns the default provider (claude).
func resolveAgents(flagVal string, w io.Writer, r io.Reader, isTTY bool) ([]string, error) {
	if flagVal != "" {
		return scaffold.ParseAgents(flagVal), nil
	}
	providers := []string{"claude", "codex", "gemini"}
	return prompt.SelectMulti(w, r, isTTY, "Select providers (toggle number, Enter to confirm):", providers, []int{0})
}

func init() {
	initCmd.Flags().StringVar(&agents, "agents", "", "comma-separated list of agents")
	rootCmd.AddCommand(initCmd)
}
