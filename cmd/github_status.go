package cmd

import "github.com/spf13/cobra"

func newGitHubStatusCmd(runner *GitHubAuthStatusRunner) *cobra.Command {
	if runner == nil {
		runner = newGitHubAuthStatusRunner(nil)
	}

	return &cobra.Command{
		Use:          "status",
		Short:        "Explain the current GitHub authentication state",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runner.Run(cmd.Context(), cmd.OutOrStdout())
		},
	}
}
