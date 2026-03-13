package cmd

import "github.com/spf13/cobra"

func newGitHubLoginCmd(runner *GitHubLoginRunner) *cobra.Command {
	if runner == nil {
		runner = newGitHubLoginRunner(nil)
	}

	return &cobra.Command{
		Use:          "login",
		Short:        "Log in to GitHub with device flow",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runner.Run(cmd.Context(), cmd.OutOrStdout())
		},
	}
}
