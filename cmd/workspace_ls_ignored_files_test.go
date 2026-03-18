package cmd

import (
	"bytes"
	"io"
	"testing"
)

func TestWorkspaceLSIgnoredFilesCommandUsesRunnerOutput(t *testing.T) {
	runner := &stubWorkspaceIgnoredFilesCommandRunner{
		run: func(stdout io.Writer) error {
			_, err := io.WriteString(stdout, "[alpha] node_modules/\n")
			return err
		},
	}

	cmd := newWorkspaceLSIgnoredFilesCmd(runner)
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	want := "[alpha] node_modules/\n"
	if output.String() != want {
		t.Fatalf("output = %q, want %q", output.String(), want)
	}
	if runner.runCount != 1 {
		t.Fatalf("runCount = %d, want %d", runner.runCount, 1)
	}
}

func TestWorkspaceCommandIncludesLSIgnoredFilesSubcommand(t *testing.T) {
	cmd := newWorkspaceCmd(newWorkspaceStatusTestInitializer(t, t.TempDir()))

	found := false
	for _, subcommand := range cmd.Commands() {
		if subcommand.Name() == "ls-ignored-files" {
			found = true
			break
		}
	}

	if !found {
		t.Fatal("expected workspace command to include the ls-ignored-files subcommand")
	}
}

type stubWorkspaceIgnoredFilesCommandRunner struct {
	run      func(stdout io.Writer) error
	runCount int
}

func (s *stubWorkspaceIgnoredFilesCommandRunner) Run(stdout io.Writer) error {
	s.runCount++
	if s.run == nil {
		return nil
	}

	return s.run(stdout)
}
