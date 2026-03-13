package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newConfigGetCmd(initializer *ConfigInitializer) *cobra.Command {
	if initializer == nil {
		initializer = newDefaultConfigInitializer()
	}

	return &cobra.Command{
		Use:          "get KEY",
		Short:        "Get a config value by dot path",
		Args:         requireSingleConfigKeyArg,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			value, err := initializer.GetValue(args[0])
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), value)
			return err
		},
	}
}

func requireSingleConfigKeyArg(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("%s requires exactly 1 argument: KEY", cmd.CommandPath())
	}

	return nil
}
