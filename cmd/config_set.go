package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newConfigSetCmd(initializer *ConfigInitializer) *cobra.Command {
	if initializer == nil {
		initializer = newDefaultConfigInitializer()
	}

	return &cobra.Command{
		Use:          "set KEY VALUE",
		Short:        "Set a config value by dot path",
		Args:         requireConfigKeyAndValueArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := initializer.SetValue(args[0], args[1]); err != nil {
				return err
			}

			storedValue, err := initializer.GetValue(args[0])
			if err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s=%s\n", args[0], storedValue)
			return err
		},
	}
}

func requireConfigKeyAndValueArgs(cmd *cobra.Command, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("%s requires exactly 2 arguments: KEY VALUE", cmd.CommandPath())
	}

	return nil
}
