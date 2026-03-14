package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newWorkspaceInitCmd(initializer *WorkspaceInitializer) *cobra.Command {
	if initializer == nil {
		initializer = newDefaultWorkspaceInitializer(nil)
	}

	return &cobra.Command{
		Use:          "init",
		Short:        "Initialize the workspace directory",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			workspacePath, err := initializer.Init()
			if err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "initialized workspace at %s\n", workspacePath)
			return err
		},
	}
}
