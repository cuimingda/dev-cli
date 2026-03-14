package cmd

import "github.com/spf13/cobra"

func newWorkspaceCmd(initializer *ConfigInitializer) *cobra.Command {
	if initializer == nil {
		initializer = newDefaultConfigInitializer()
	}

	workspaceInitializer := newDefaultWorkspaceInitializer(initializer)
	workspaceLister := newDefaultWorkspaceLister(initializer)

	cmd := &cobra.Command{
		Use:   "workspace",
		Short: "Manage workspaces",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newWorkspaceInitCmd(workspaceInitializer))
	cmd.AddCommand(newWorkspaceListCmd(workspaceLister))

	return cmd
}
