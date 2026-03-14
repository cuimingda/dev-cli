package cmd

import "github.com/spf13/cobra"

func newConfigResolvedCmd(initializer *ConfigInitializer) *cobra.Command {
	if initializer == nil {
		initializer = newDefaultConfigInitializer()
	}

	return &cobra.Command{
		Use:          "resolved",
		Short:        "List resolved config values as key=value pairs",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			entries, err := initializer.ListResolvedKeyValues()
			if err != nil {
				return err
			}

			return writeConfigEntries(cmd, entries)
		},
	}
}
