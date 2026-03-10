package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCommandShowsHelpWhenRunWithoutArgs(t *testing.T) {
	cmd := newRootCmd()
	var output bytes.Buffer

	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs(nil)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	help := output.String()
	if !strings.Contains(help, "Usage:") {
		t.Fatalf("expected help output to contain Usage, got %q", help)
	}

	if !strings.Contains(help, "dev [flags]") {
		t.Fatalf("expected help output to contain root usage, got %q", help)
	}
}

func TestExecuteReturnsErrorWhenNotOnMacOS(t *testing.T) {
	originalGOOS := currentGOOS
	originalRootCmd := rootCmd
	t.Cleanup(func() {
		currentGOOS = originalGOOS
		rootCmd = originalRootCmd
	})

	currentGOOS = func() string {
		return "linux"
	}
	rootCmd = newRootCmd()

	err := Execute()
	if err == nil {
		t.Fatal("expected Execute() to return an error on non-macOS")
	}

	if !strings.Contains(err.Error(), "only supports macOS") {
		t.Fatalf("expected macOS restriction error, got %q", err.Error())
	}
}

func TestExecuteShowsHelpOnMacOS(t *testing.T) {
	originalGOOS := currentGOOS
	originalRootCmd := rootCmd
	t.Cleanup(func() {
		currentGOOS = originalGOOS
		rootCmd = originalRootCmd
	})

	currentGOOS = func() string {
		return "darwin"
	}
	rootCmd = newRootCmd()

	var output bytes.Buffer
	rootCmd.SetOut(&output)
	rootCmd.SetErr(&output)
	rootCmd.SetArgs(nil)

	if err := Execute(); err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	help := output.String()
	if !strings.Contains(help, "Usage:") {
		t.Fatalf("expected help output to contain Usage, got %q", help)
	}
}
