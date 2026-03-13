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
		Short:        "List config keys as dot paths",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			paths, err := initializer.ListDotPaths()
			if err != nil {
				return err
			}

			for _, path := range paths {
				if _, err := fmt.Fprintln(cmd.OutOrStdout(), path); err != nil {
					return err
				}
			}

			return nil
		},
	}
}
