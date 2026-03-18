package cmd

import (
	"bytes"
	"io"
	"reflect"
	"testing"

	"github.com/spf13/cobra"
)

func TestCDCommandUsesRunnerOutput(t *testing.T) {
	runner := &stubWorkspaceCDCommandRunner{
		run: func(stdout io.Writer, projectName string) error {
			if projectName != "alpha" {
				t.Fatalf("projectName = %q, want %q", projectName, "alpha")
			}

			_, err := io.WriteString(stdout, "/tmp/workspace/alpha\n")
			return err
		},
	}

	cmd := newCDCmd(runner)
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs([]string{"alpha"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	want := "/tmp/workspace/alpha\n"
	if output.String() != want {
		t.Fatalf("output = %q, want %q", output.String(), want)
	}
	if runner.runCount != 1 {
		t.Fatalf("runCount = %d, want %d", runner.runCount, 1)
	}
}

func TestCDCommandRequiresProjectArg(t *testing.T) {
	cmd := newCDCmd(&stubWorkspaceCDCommandRunner{})
	cmd.SetArgs(nil)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected Execute() to return an error when no project name is provided")
	}
}

func TestCDCommandCompletionUsesRunner(t *testing.T) {
	runner := &stubWorkspaceCDCommandRunner{
		complete: func(toComplete string) ([]string, cobra.ShellCompDirective, error) {
			if toComplete != "al" {
				t.Fatalf("toComplete = %q, want %q", toComplete, "al")
			}

			return []string{"alpha"}, cobra.ShellCompDirectiveNoFileComp, nil
		},
	}

	cmd := newCDCmd(runner)
	completions, directive := cmd.ValidArgsFunction(cmd, nil, "al")

	if !reflect.DeepEqual(completions, []string{"alpha"}) {
		t.Fatalf("completions = %#v, want %#v", completions, []string{"alpha"})
	}
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("directive = %v, want %v", directive, cobra.ShellCompDirectiveNoFileComp)
	}
	if runner.completeCount != 1 {
		t.Fatalf("completeCount = %d, want %d", runner.completeCount, 1)
	}
}

func TestRootCommandIncludesCDCommand(t *testing.T) {
	cmd := newRootCmdWithConfigInitializer(newWorkspaceStatusTestInitializer(t, t.TempDir()))

	found := false
	for _, subcommand := range cmd.Commands() {
		if subcommand.Name() == "cd" {
			found = true
			break
		}
	}

	if !found {
		t.Fatal("expected root command to include the cd command")
	}
}

type stubWorkspaceCDCommandRunner struct {
	run           func(stdout io.Writer, projectName string) error
	complete      func(toComplete string) ([]string, cobra.ShellCompDirective, error)
	runCount      int
	completeCount int
}

func (s *stubWorkspaceCDCommandRunner) Run(stdout io.Writer, projectName string) error {
	s.runCount++
	if s.run == nil {
		return nil
	}

	return s.run(stdout, projectName)
}

func (s *stubWorkspaceCDCommandRunner) Complete(toComplete string) ([]string, cobra.ShellCompDirective, error) {
	s.completeCount++
	if s.complete == nil {
		return nil, cobra.ShellCompDirectiveNoFileComp, nil
	}

	return s.complete(toComplete)
}
