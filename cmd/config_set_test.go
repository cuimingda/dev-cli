package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	configtemplate "github.com/cuimingda/dev-cli/config"
)

func TestConfigInitializerSetValueUpdatesExistingKey(t *testing.T) {
	initializer := &ConfigInitializer{
		configHome:   t.TempDir(),
		templateYAML: configtemplate.TemplateYAML(),
	}

	if _, err := initializer.Init(); err != nil {
		t.Fatalf("Init() returned error: %v", err)
	}

	if err := initializer.SetValue("github.client_id", "abc123"); err != nil {
		t.Fatalf("SetValue() returned error: %v", err)
	}

	value, err := initializer.GetValue("github.client_id")
	if err != nil {
		t.Fatalf("GetValue() returned error: %v", err)
	}

	if value != "abc123" {
		t.Fatalf("GetValue() = %q, want %q", value, "abc123")
	}
}

func TestConfigInitializerSetValueCreatesNestedKey(t *testing.T) {
	initializer := &ConfigInitializer{
		configHome:   t.TempDir(),
		templateYAML: configtemplate.TemplateYAML(),
	}

	if _, err := initializer.Init(); err != nil {
		t.Fatalf("Init() returned error: %v", err)
	}

	if err := initializer.SetValue("github.nested.callback_url", "https://example.com/callback"); err != nil {
		t.Fatalf("SetValue() returned error: %v", err)
	}

	value, err := initializer.GetValue("github.nested.callback_url")
	if err != nil {
		t.Fatalf("GetValue() returned error: %v", err)
	}

	if value != "https://example.com/callback" {
		t.Fatalf("GetValue() = %q, want %q", value, "https://example.com/callback")
	}
}

func TestConfigInitializerSetValueExpandsWorkspaceRootHomeReferences(t *testing.T) {
	t.Setenv("HOME", "/tmp/dev-cli-home")

	initializer := &ConfigInitializer{
		configHome:   t.TempDir(),
		templateYAML: configtemplate.TemplateYAML(),
	}

	if _, err := initializer.Init(); err != nil {
		t.Fatalf("Init() returned error: %v", err)
	}

	testCases := []struct {
		name  string
		value string
		want  string
	}{
		{
			name:  "dollar home",
			value: "$HOME/Projects",
			want:  "/tmp/dev-cli-home/Projects",
		},
		{
			name:  "tilde",
			value: "~/Projects",
			want:  "/tmp/dev-cli-home/Projects",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if err := initializer.SetValue("workspace.root", testCase.value); err != nil {
				t.Fatalf("SetValue() returned error: %v", err)
			}

			value, err := initializer.GetValue("workspace.root")
			if err != nil {
				t.Fatalf("GetValue() returned error: %v", err)
			}
			if value != testCase.want {
				t.Fatalf("GetValue() = %q, want %q", value, testCase.want)
			}

			content, err := os.ReadFile(initializer.DefaultPath())
			if err != nil {
				t.Fatalf("ReadFile() returned error: %v", err)
			}
			if !strings.Contains(string(content), testCase.want) {
				t.Fatalf("config content = %q, want it to contain %q", string(content), testCase.want)
			}
		})
	}
}

func TestConfigInitializerSetValueReturnsErrorWhenParentIsScalar(t *testing.T) {
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
		"  api_base_url: \"https://api.github.com\"",
		"",
	}, "\n")
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("WriteFile() returned error: %v", err)
	}

	err := initializer.SetValue("github.api_base_url.host", "example.com")
	if err == nil {
		t.Fatal("expected SetValue() to return an error when parent is scalar")
	}

	if !strings.Contains(err.Error(), "parent is not a mapping") {
		t.Fatalf("expected scalar parent error, got %q", err.Error())
	}
}

func TestConfigSetCommandWritesValue(t *testing.T) {
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
	cmd.SetArgs([]string{"config", "set", "github.client_id", "abc123"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	if output.String() != "github.client_id=abc123\n" {
		t.Fatalf("output = %q, want %q", output.String(), "github.client_id=abc123\n")
	}

	value, err := initializer.GetValue("github.client_id")
	if err != nil {
		t.Fatalf("GetValue() returned error: %v", err)
	}

	if value != "abc123" {
		t.Fatalf("GetValue() = %q, want %q", value, "abc123")
	}
}

func TestConfigSetCommandPrintsNormalizedWorkspaceRoot(t *testing.T) {
	t.Setenv("HOME", "/tmp/dev-cli-home")

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
	cmd.SetArgs([]string{"config", "set", "workspace.root", "~/Projects"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	if output.String() != "workspace.root=/tmp/dev-cli-home/Projects\n" {
		t.Fatalf("output = %q, want %q", output.String(), "workspace.root=/tmp/dev-cli-home/Projects\n")
	}

	value, err := initializer.GetValue("workspace.root")
	if err != nil {
		t.Fatalf("GetValue() returned error: %v", err)
	}
	if value != "/tmp/dev-cli-home/Projects" {
		t.Fatalf("GetValue() = %q, want %q", value, "/tmp/dev-cli-home/Projects")
	}
}

func TestConfigSetCommandRequiresExactlyTwoArgs(t *testing.T) {
	initializer := &ConfigInitializer{
		configHome:   t.TempDir(),
		templateYAML: configtemplate.TemplateYAML(),
	}

	cmd := newRootCmdWithConfigInitializer(initializer)
	var output bytes.Buffer

	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs([]string{"config", "set", "github.client_id"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected Execute() to return an error when value argument is missing")
	}

	if !strings.Contains(err.Error(), "dev config set requires exactly 2 arguments: KEY VALUE") {
		t.Fatalf("expected argument error, got %q", err.Error())
	}

	if output.String() != "" {
		t.Fatalf("expected command output to be empty on error, got %q", output.String())
	}
}
