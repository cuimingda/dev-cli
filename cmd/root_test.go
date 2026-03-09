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
