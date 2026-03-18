package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/spf13/cobra"
)

func TestWorkspaceCDRunnerRunPrintsResolvedProjectPath(t *testing.T) {
	workspaceRoot := t.TempDir()
	createWorkspaceProject(t, workspaceRoot, "alpha", `[remote "origin"]`+"\n\turl = git@github.com:openai/alpha.git\n")

	runner := &WorkspaceCDRunner{
		configInitializer: newWorkspaceStatusTestInitializer(t, workspaceRoot),
	}

	var output bytes.Buffer
	if err := runner.Run(&output, "alpha"); err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}

	want := filepath.Join(workspaceRoot, "alpha") + "\n"
	if output.String() != want {
		t.Fatalf("output = %q, want %q", output.String(), want)
	}
}

func TestWorkspaceCDRunnerResolveRejectsTraversal(t *testing.T) {
	runner := &WorkspaceCDRunner{
		configInitializer: newWorkspaceStatusTestInitializer(t, t.TempDir()),
	}

	_, err := runner.Resolve("../alpha")
	if err == nil {
		t.Fatal("expected Resolve() to return an error for a non-directory-name argument")
	}
	if got := err.Error(); got != "workspace project name must be a single directory name" {
		t.Fatalf("error = %q, want %q", got, "workspace project name must be a single directory name")
	}
}

func TestWorkspaceCDRunnerCompleteReturnsWorkspaceProjectNames(t *testing.T) {
	workspaceRoot := t.TempDir()
	createWorkspaceProject(t, workspaceRoot, "alpha", `[remote "origin"]`+"\n\turl = git@github.com:openai/alpha.git\n")
	createWorkspaceProject(t, workspaceRoot, "beta", `[remote "origin"]`+"\n\turl = git@github.com:openai/beta.git\n")
	if err := os.WriteFile(filepath.Join(workspaceRoot, "notes.txt"), []byte("ignore"), 0o644); err != nil {
		t.Fatalf("WriteFile() returned error: %v", err)
	}

	runner := &WorkspaceCDRunner{
		configInitializer: newWorkspaceStatusTestInitializer(t, workspaceRoot),
	}

	completions, directive, err := runner.Complete("a")
	if err != nil {
		t.Fatalf("Complete() returned error: %v", err)
	}

	wantCompletions := []string{"alpha"}
	if !reflect.DeepEqual(completions, wantCompletions) {
		t.Fatalf("completions = %#v, want %#v", completions, wantCompletions)
	}
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("directive = %v, want %v", directive, cobra.ShellCompDirectiveNoFileComp)
	}
}
