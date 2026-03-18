package cmd

import (
	"io"

	"github.com/spf13/cobra"
)

type workspaceIgnoredFilesCommandRunner interface {
	Run(stdout io.Writer) error
}

func newWorkspaceLSIgnoredFilesCmd(runner workspaceIgnoredFilesCommandRunner) *cobra.Command {
	if runner == nil {
		runner = newDefaultWorkspaceIgnoredFilesRunner(nil)
	}

	return &cobra.Command{
		Use:          "ls-ignored-files",
		Short:        "List ignored, untracked files for each workspace project",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runner.Run(cmd.OutOrStdout())
		},
	}
}
