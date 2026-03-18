package cmd

import (
	"io"

	"github.com/spf13/cobra"
)

type workspaceCDCommandRunner interface {
	Run(stdout io.Writer, projectName string) error
	Complete(toComplete string) ([]string, cobra.ShellCompDirective, error)
}

func newCDCmd(runner workspaceCDCommandRunner) *cobra.Command {
	if runner == nil {
		runner = newDefaultWorkspaceCDRunner(nil)
	}

	cmd := &cobra.Command{
		Use:          "cd NAME",
		Short:        "Resolve a workspace project path for shell cd integration",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runner.Run(cmd.OutOrStdout(), args[0])
		},
	}

	cmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		completions, directive, err := runner.Complete(toComplete)
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		return completions, directive
	}

	return cmd
}
