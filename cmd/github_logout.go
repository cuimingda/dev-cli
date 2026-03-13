package cmd

import "github.com/spf13/cobra"

func newGitHubLogoutCmd(runner *GitHubLogoutRunner) *cobra.Command {
	if runner == nil {
		runner = newGitHubLogoutRunner(nil)
	}

	return &cobra.Command{
		Use:          "logout",
		Short:        "Log out from GitHub on this machine",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runner.Run(cmd.Context(), cmd.OutOrStdout())
		},
	}
}
