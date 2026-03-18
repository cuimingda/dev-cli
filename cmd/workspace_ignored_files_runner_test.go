package cmd

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestWorkspaceIgnoredFilesRunnerListReturnsIgnoredPathsAcrossWorkspaceProjects(t *testing.T) {
	workspaceRoot := t.TempDir()
	createWorkspaceProject(t, workspaceRoot, "beta", `[remote "origin"]`+"\n\turl = git@github.com:openai/beta.git\n")
	createWorkspaceProject(t, workspaceRoot, "alpha", `[remote "origin"]`+"\n\turl = git@github.com:openai/alpha.git\n")
	createWorkspaceProject(t, workspaceRoot, "delta", `[core]`+"\n\trepositoryformatversion = 0\n")
	if err := os.MkdirAll(filepath.Join(workspaceRoot, "gamma"), 0o755); err != nil {
		t.Fatalf("MkdirAll() returned error: %v", err)
	}

	initializer := newWorkspaceStatusTestInitializer(t, workspaceRoot)
	gitCalls := []string{}
	runner := &WorkspaceIgnoredFilesRunner{
		configInitializer: initializer,
		runGit: func(projectPath string, args ...string) ([]byte, error) {
			call := filepath.Base(projectPath) + ":" + strings.Join(args, " ")
			gitCalls = append(gitCalls, call)

			switch call {
			case "alpha:ls-files --others --ignored --exclude-standard --directory --no-empty-directory -z":
				return []byte("node_modules/\x00build/output.log\x00"), nil
			case "beta:ls-files --others --ignored --exclude-standard --directory --no-empty-directory -z":
				return []byte("tmp/debug.log\x00tmp/cache/\x00"), nil
			case "delta:ls-files --others --ignored --exclude-standard --directory --no-empty-directory -z":
				return []byte(""), nil
			default:
				t.Fatalf("unexpected git invocation: %s", call)
				return nil, errors.New("unexpected git invocation")
			}
		},
	}

	entries, err := runner.List()
	if err != nil {
		t.Fatalf("List() returned error: %v", err)
	}

	wantEntries := []WorkspaceIgnoredFileEntry{
		{Project: "alpha", Path: "build/output.log"},
		{Project: "alpha", Path: "node_modules/"},
		{Project: "beta", Path: "tmp/cache/"},
		{Project: "beta", Path: "tmp/debug.log"},
	}
	if !reflect.DeepEqual(entries, wantEntries) {
		t.Fatalf("List() = %#v, want %#v", entries, wantEntries)
	}

	wantGitCalls := []string{
		"alpha:ls-files --others --ignored --exclude-standard --directory --no-empty-directory -z",
		"beta:ls-files --others --ignored --exclude-standard --directory --no-empty-directory -z",
		"delta:ls-files --others --ignored --exclude-standard --directory --no-empty-directory -z",
	}
	if !reflect.DeepEqual(gitCalls, wantGitCalls) {
		t.Fatalf("git calls = %#v, want %#v", gitCalls, wantGitCalls)
	}
}

func TestWorkspaceIgnoredFilesRunnerRunFormatsEntries(t *testing.T) {
	workspaceRoot := t.TempDir()
	createWorkspaceProject(t, workspaceRoot, "alpha", `[remote "origin"]`+"\n\turl = git@github.com:openai/alpha.git\n")

	runner := &WorkspaceIgnoredFilesRunner{
		configInitializer: newWorkspaceStatusTestInitializer(t, workspaceRoot),
		runGit: func(projectPath string, args ...string) ([]byte, error) {
			return []byte("node_modules/\x00build/output.log\x00"), nil
		},
	}

	var output bytes.Buffer
	if err := runner.Run(&output); err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}

	want := strings.Join([]string{
		"[alpha] build/output.log",
		"[alpha] node_modules/",
		"",
	}, "\n")
	if output.String() != want {
		t.Fatalf("output = %q, want %q", output.String(), want)
	}
}

func TestWorkspaceIgnoredFilesRunnerListReturnsErrorWhenWorkspaceRootDoesNotExist(t *testing.T) {
	runner := &WorkspaceIgnoredFilesRunner{
		configInitializer: newWorkspaceStatusTestInitializer(t, filepath.Join(t.TempDir(), "missing")),
		runGit: func(projectPath string, args ...string) ([]byte, error) {
			t.Fatalf("unexpected git invocation for %s %v", projectPath, args)
			return nil, nil
		},
	}

	_, err := runner.List()
	if err == nil {
		t.Fatal("expected List() to return an error when workspace root does not exist")
	}
}
