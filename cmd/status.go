package cmd

import "github.com/spf13/cobra"

func newStatusCmd(runner workspaceStatusCommandRunner) *cobra.Command {
	if runner == nil {
		runner = newDefaultWorkspaceStatusRunner(nil)
	}

	return &cobra.Command{
		Use:          "status",
		Short:        "Alias for workspace status",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runner.Run(cmd.OutOrStdout())
		},
	}
}
