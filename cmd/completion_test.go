package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestCompletionCommandShowsHelpWithoutArgs(t *testing.T) {
	cmd := newCompletionCmd(newRootCmdWithConfigInitializer(newWorkspaceStatusTestInitializer(t, t.TempDir())))
	var output bytes.Buffer

	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs(nil)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	if !strings.Contains(output.String(), "Available Commands:") {
		t.Fatalf("output = %q, want completion help", output.String())
	}
}

func TestCompletionZshCommandOutputsShellIntegration(t *testing.T) {
	cmd := newRootCmdWithConfigInitializer(newWorkspaceStatusTestInitializer(t, t.TempDir()))
	var output bytes.Buffer

	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs([]string{"completion", "zsh"})

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

func TestRootCommandIncludesCompletionCommand(t *testing.T) {
	cmd := newRootCmdWithConfigInitializer(newWorkspaceStatusTestInitializer(t, t.TempDir()))

	found := false
	for _, subcommand := range cmd.Commands() {
		if subcommand.Name() == "completion" {
			found = true
			break
		}
	}

	if !found {
		t.Fatal("expected root command to include the completion command")
	}
}

func TestRootCommandDoesNotIncludeZshInitCommand(t *testing.T) {
	cmd := newRootCmdWithConfigInitializer(newWorkspaceStatusTestInitializer(t, t.TempDir()))

	for _, subcommand := range cmd.Commands() {
		if subcommand.Name() == "zsh-init" {
			t.Fatal("did not expect root command to include zsh-init")
		}
	}
}
