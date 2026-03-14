package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	configtemplate "github.com/cuimingda/dev-cli/config"
)

func TestWorkspaceInitCommandCreatesDirectory(t *testing.T) {
	workspaceRoot := filepath.Join(t.TempDir(), "workspace")

	configInitializer := &ConfigInitializer{
		configHome:   t.TempDir(),
		templateYAML: configtemplate.TemplateYAML(),
		defaultYAML:  configtemplate.DefaultYAML(),
	}
	if _, err := configInitializer.Init(); err != nil {
		t.Fatalf("Init() returned error: %v", err)
	}
	if err := configInitializer.SetValue("workspace.root", workspaceRoot); err != nil {
		t.Fatalf("SetValue() returned error: %v", err)
	}

	cmd := newRootCmdWithConfigInitializer(configInitializer)
	var output bytes.Buffer

	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs([]string{"workspace", "init"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	if output.String() != "initialized workspace at "+workspaceRoot+"\n" {
		t.Fatalf("output = %q, want %q", output.String(), "initialized workspace at "+workspaceRoot+"\n")
	}

	info, err := os.Stat(workspaceRoot)
	if err != nil {
		t.Fatalf("Stat() returned error: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("expected %q to be a directory", workspaceRoot)
	}
}

func TestWorkspaceInitCommandReturnsErrorWhenDirectoryExists(t *testing.T) {
	workspaceRoot := filepath.Join(t.TempDir(), "workspace")
	if err := os.MkdirAll(workspaceRoot, 0o755); err != nil {
		t.Fatalf("MkdirAll() returned error: %v", err)
	}

	configInitializer := &ConfigInitializer{
		configHome:   t.TempDir(),
		templateYAML: configtemplate.TemplateYAML(),
		defaultYAML:  configtemplate.DefaultYAML(),
	}
	if _, err := configInitializer.Init(); err != nil {
		t.Fatalf("Init() returned error: %v", err)
	}
	if err := configInitializer.SetValue("workspace.root", workspaceRoot); err != nil {
		t.Fatalf("SetValue() returned error: %v", err)
	}

	cmd := newRootCmdWithConfigInitializer(configInitializer)
	var output bytes.Buffer

	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs([]string{"workspace", "init"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected Execute() to return an error when workspace directory exists")
	}

	if !strings.Contains(err.Error(), "workspace directory already exists") {
		t.Fatalf("expected already exists error, got %q", err.Error())
	}

	if output.String() != "" {
		t.Fatalf("expected command output to be empty on error, got %q", output.String())
	}
}
