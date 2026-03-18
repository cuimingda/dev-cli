package cmd

import (
	"bytes"
	"io"
	"testing"
)

func TestStatusCommandUsesRunnerOutput(t *testing.T) {
	runner := &stubWorkspaceStatusCommandRunner{
		run: func(stdout io.Writer) error {
			_, err := io.WriteString(stdout, "alpha - git: ✅ remote: ✅ clean: ✅ synced: ✅\n")
			return err
		},
	}

	cmd := newStatusCmd(runner)
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	want := "alpha - git: ✅ remote: ✅ clean: ✅ synced: ✅\n"
	if output.String() != want {
		t.Fatalf("output = %q, want %q", output.String(), want)
	}
	if runner.runCount != 1 {
		t.Fatalf("runCount = %d, want %d", runner.runCount, 1)
	}
}

func TestRootCommandIncludesStatusAlias(t *testing.T) {
	cmd := newRootCmdWithConfigInitializer(newWorkspaceStatusTestInitializer(t, t.TempDir()))

	found := false
	for _, subcommand := range cmd.Commands() {
		if subcommand.Name() == "status" {
			found = true
			break
		}
	}

	if !found {
		t.Fatal("expected root command to include the status alias")
	}
}
