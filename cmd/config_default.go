package cmd

import "github.com/spf13/cobra"

func newConfigDefaultCmd(initializer *ConfigInitializer) *cobra.Command {
	if initializer == nil {
		initializer = newDefaultConfigInitializer()
	}

	return &cobra.Command{
		Use:          "default",
		Short:        "List default config values as key=value pairs",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			entries, err := initializer.ListDefaultKeyValues()
			if err != nil {
				return err
			}

			return writeConfigEntries(cmd, entries)
		},
	}
}
