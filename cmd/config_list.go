package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newConfigListCmd(initializer *ConfigInitializer) *cobra.Command {
	if initializer == nil {
		initializer = newDefaultConfigInitializer()
	}

	return &cobra.Command{
		Use:          "list",
		Short:        "List config values as key=value pairs",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			entries, err := initializer.ListKeyValues()
			if err != nil {
				return err
			}

			for _, entry := range entries {
				if _, err := fmt.Fprintln(cmd.OutOrStdout(), entry); err != nil {
					return err
				}
			}

			return nil
		},
	}
}
