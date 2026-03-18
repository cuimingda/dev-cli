package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

func newZshInitCmd(rootCmd *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:          "zsh-init",
		Short:        "Print zsh integration for dev cd and completion",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return writeZshInitScript(cmd.OutOrStdout(), rootCmd)
		},
	}
}

func writeZshInitScript(stdout io.Writer, rootCmd *cobra.Command) error {
	if stdout == nil {
		stdout = io.Discard
	}
	if rootCmd == nil {
		return fmt.Errorf("root command is not configured")
	}

	if _, err := fmt.Fprintln(stdout, "function dev() {"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(stdout, "  if [[ \"$1\" == \"cd\" ]]; then"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(stdout, "    shift"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(stdout, "    local target"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(stdout, "    target=$(command dev cd \"$@\") || return $?"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(stdout, "    builtin cd -- \"$target\""); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(stdout, "  else"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(stdout, "    command dev \"$@\""); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(stdout, "  fi"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(stdout, "}"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(stdout, "if ! whence compdef >/dev/null 2>&1; then"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(stdout, "  autoload -Uz compinit"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(stdout, "  compinit"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(stdout, "fi"); err != nil {
		return err
	}

	return rootCmd.GenZshCompletion(stdout)
}
