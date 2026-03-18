package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestZshInitCommandOutputsShellIntegration(t *testing.T) {
	cmd := newRootCmdWithConfigInitializer(newWorkspaceStatusTestInitializer(t, t.TempDir()))
	var output bytes.Buffer

	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs([]string{"zsh-init"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	script := output.String()
	if !strings.Contains(script, "function dev() {") {
		t.Fatalf("output = %q, want zsh wrapper function", script)
	}
	if !strings.Contains(script, "target=$(command dev cd \"$@\") || return $?") {
		t.Fatalf("output = %q, want cd wrapper integration", script)
	}
	if !strings.Contains(script, "compdef _dev dev") {
		t.Fatalf("output = %q, want zsh completion registration", script)
	}
}

func TestRootCommandIncludesZshInitCommand(t *testing.T) {
	cmd := newRootCmdWithConfigInitializer(newWorkspaceStatusTestInitializer(t, t.TempDir()))

	found := false
	for _, subcommand := range cmd.Commands() {
		if subcommand.Name() == "zsh-init" {
			found = true
			break
		}
	}

	if !found {
		t.Fatal("expected root command to include the zsh-init command")
	}
}
