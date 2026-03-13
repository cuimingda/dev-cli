package cmd

import "github.com/spf13/cobra"

func newConfigUnsetCmd(initializer *ConfigInitializer) *cobra.Command {
	if initializer == nil {
		initializer = newDefaultConfigInitializer()
	}

	return &cobra.Command{
		Use:          "unset KEY",
		Short:        "Delete a config value by dot path",
		Args:         requireSingleConfigKeyArg,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return initializer.UnsetValue(args[0])
		},
	}
}
