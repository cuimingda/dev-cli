package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	configtemplate "github.com/cuimingda/dev-cli/config"
)

func TestConfigInitializerGetValue(t *testing.T) {
	initializer := &ConfigInitializer{
		configHome:   t.TempDir(),
		templateYAML: configtemplate.TemplateYAML(),
	}

	configPath := initializer.DefaultPath()
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() returned error: %v", err)
	}

	configContent := strings.Join([]string{
		"github:",
		"  client_id:",
		"  api_base_url: \"https://api.github.com\"",
		"  nested:",
		"    callback_url: \"https://example.com/callback\"",
		"",
	}, "\n")

	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("WriteFile() returned error: %v", err)
	}

	apiBaseURL, err := initializer.GetValue("github.api_base_url")
	if err != nil {
		t.Fatalf("GetValue() returned error: %v", err)
	}

	if apiBaseURL != "https://api.github.com" {
		t.Fatalf("GetValue() = %q, want %q", apiBaseURL, "https://api.github.com")
	}

	clientID, err := initializer.GetValue("github.client_id")
	if err != nil {
		t.Fatalf("GetValue() returned error for empty value: %v", err)
	}

	if clientID != "" {
		t.Fatalf("GetValue() = %q, want empty string", clientID)
	}
}

func TestConfigGetCommandPrintsValue(t *testing.T) {
	initializer := &ConfigInitializer{
		configHome:   t.TempDir(),
		templateYAML: configtemplate.TemplateYAML(),
	}

	if _, err := initializer.Init(); err != nil {
		t.Fatalf("Init() returned error: %v", err)
	}

	cmd := newRootCmdWithConfigInitializer(initializer)
	var output bytes.Buffer

	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs([]string{"config", "get", "github.api_base_url"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	if output.String() != "https://api.github.com\n" {
		t.Fatalf("output = %q, want %q", output.String(), "https://api.github.com\n")
	}
}

func TestConfigGetCommandPrintsEmptyValue(t *testing.T) {
	initializer := &ConfigInitializer{
		configHome:   t.TempDir(),
		templateYAML: configtemplate.TemplateYAML(),
	}

	if _, err := initializer.Init(); err != nil {
		t.Fatalf("Init() returned error: %v", err)
	}

	cmd := newRootCmdWithConfigInitializer(initializer)
	var output bytes.Buffer

	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs([]string{"config", "get", "github.client_id"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	if output.String() != "\n" {
		t.Fatalf("output = %q, want newline only", output.String())
	}
}

func TestConfigGetCommandReturnsErrorWhenKeyDoesNotExist(t *testing.T) {
	initializer := &ConfigInitializer{
		configHome:   t.TempDir(),
		templateYAML: configtemplate.TemplateYAML(),
	}

	if _, err := initializer.Init(); err != nil {
		t.Fatalf("Init() returned error: %v", err)
	}

	cmd := newRootCmdWithConfigInitializer(initializer)
	var output bytes.Buffer

	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs([]string{"config", "get", "github.unknown"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected Execute() to return an error for missing config key")
	}

	if !strings.Contains(err.Error(), "config key not found") {
		t.Fatalf("expected missing config key error, got %q", err.Error())
	}

	if output.String() != "" {
		t.Fatalf("expected command output to be empty on error, got %q", output.String())
	}
}

func TestConfigGetCommandRequiresExactlyOneArg(t *testing.T) {
	initializer := &ConfigInitializer{
		configHome:   t.TempDir(),
		templateYAML: configtemplate.TemplateYAML(),
	}

	cmd := newRootCmdWithConfigInitializer(initializer)
	var output bytes.Buffer

	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs([]string{"config", "get"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected Execute() to return an error when key argument is missing")
	}

	if !strings.Contains(err.Error(), "dev config get requires exactly 1 argument: KEY") {
		t.Fatalf("expected argument error, got %q", err.Error())
	}

	if output.String() != "" {
		t.Fatalf("expected command output to be empty on error, got %q", output.String())
	}
}
