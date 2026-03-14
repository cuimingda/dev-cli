package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	configtemplate "github.com/cuimingda/dev-cli/config"
)

func TestWorkspaceInitializerInitCreatesDirectoryFromResolvedConfig(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	configInitializer := &ConfigInitializer{
		configHome:   t.TempDir(),
		templateYAML: configtemplate.TemplateYAML(),
		defaultYAML:  configtemplate.DefaultYAML(),
	}
	initializer := &WorkspaceInitializer{
		configInitializer: configInitializer,
	}

	workspacePath, err := initializer.Init()
	if err != nil {
		t.Fatalf("Init() returned error: %v", err)
	}

	want := filepath.Join(homeDir, "Projects")
	if workspacePath != want {
		t.Fatalf("Init() = %q, want %q", workspacePath, want)
	}

	info, err := os.Stat(workspacePath)
	if err != nil {
		t.Fatalf("Stat() returned error: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("expected %q to be a directory", workspacePath)
	}
}

func TestWorkspaceInitializerInitReturnsErrorWhenDirectoryExists(t *testing.T) {
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

	initializer := &WorkspaceInitializer{
		configInitializer: configInitializer,
	}

	_, err := initializer.Init()
	if err == nil {
		t.Fatal("expected Init() to return an error when workspace directory exists")
	}

	if !strings.Contains(err.Error(), "workspace directory already exists") {
		t.Fatalf("expected already exists error, got %q", err.Error())
	}
}

func TestWorkspaceInitializerInitReturnsErrorWhenWorkspaceRootIsEmpty(t *testing.T) {
	configInitializer := &ConfigInitializer{
		configHome:   t.TempDir(),
		templateYAML: configtemplate.TemplateYAML(),
		defaultYAML:  configtemplate.DefaultYAML(),
	}
	if _, err := configInitializer.Init(); err != nil {
		t.Fatalf("Init() returned error: %v", err)
	}
	if err := configInitializer.SetValue("workspace.root", ""); err != nil {
		t.Fatalf("SetValue() returned error: %v", err)
	}

	initializer := &WorkspaceInitializer{
		configInitializer: configInitializer,
	}

	_, err := initializer.Init()
	if err == nil {
		t.Fatal("expected Init() to return an error when workspace.root is empty")
	}

	if got := err.Error(); got != "config value workspace.root is empty" {
		t.Fatalf("error = %q, want %q", got, "config value workspace.root is empty")
	}
}
