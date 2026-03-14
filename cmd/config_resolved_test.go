package cmd

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	configtemplate "github.com/cuimingda/dev-cli/config"
)

func TestConfigInitializerGetResolvedValueUsesDefaults(t *testing.T) {
	t.Setenv("HOME", "/tmp/dev-cli-home")

	initializer := &ConfigInitializer{
		configHome:   t.TempDir(),
		templateYAML: configtemplate.TemplateYAML(),
		defaultYAML:  configtemplate.DefaultYAML(),
	}

	apiBaseURL, err := initializer.GetResolvedValue("github.api_base_url")
	if err != nil {
		t.Fatalf("GetResolvedValue() returned error: %v", err)
	}
	if apiBaseURL != "https://api.github.com" {
		t.Fatalf("GetResolvedValue() = %q, want %q", apiBaseURL, "https://api.github.com")
	}

	clientID, err := initializer.GetResolvedValue("github.client_id")
	if err != nil {
		t.Fatalf("GetResolvedValue() returned error for empty value: %v", err)
	}
	if clientID != "" {
		t.Fatalf("GetResolvedValue() = %q, want empty string", clientID)
	}

	workspaceRoot, err := initializer.GetResolvedValue("workspace.root")
	if err != nil {
		t.Fatalf("GetResolvedValue() returned error for workspace.root: %v", err)
	}
	if workspaceRoot != "/tmp/dev-cli-home/Projects" {
		t.Fatalf("GetResolvedValue() = %q, want %q", workspaceRoot, "/tmp/dev-cli-home/Projects")
	}
}

func TestConfigInitializerListResolvedKeyValuesMergesDefaultsAndUserConfig(t *testing.T) {
	t.Setenv("HOME", "/tmp/dev-cli-home")

	initializer := &ConfigInitializer{
		configHome:   t.TempDir(),
		templateYAML: configtemplate.TemplateYAML(),
		defaultYAML:  configtemplate.DefaultYAML(),
	}

	if _, err := initializer.Init(); err != nil {
		t.Fatalf("Init() returned error: %v", err)
	}
	if err := initializer.SetValue("github.client_id", "abc123"); err != nil {
		t.Fatalf("SetValue() returned error: %v", err)
	}
	if err := initializer.SetValue("github.api_base_url", "https://ghe.example.com/api/v3"); err != nil {
		t.Fatalf("SetValue() returned error: %v", err)
	}
	if err := initializer.SetValue("github.nested.callback_url", "https://example.com/callback"); err != nil {
		t.Fatalf("SetValue() returned error: %v", err)
	}

	entries, err := initializer.ListResolvedKeyValues()
	if err != nil {
		t.Fatalf("ListResolvedKeyValues() returned error: %v", err)
	}

	want := []string{
		"github.api_base_url=https://ghe.example.com/api/v3",
		"github.client_id=abc123",
		"github.nested.callback_url=https://example.com/callback",
		"workspace.root=/tmp/dev-cli-home/Projects",
	}
	if !reflect.DeepEqual(entries, want) {
		t.Fatalf("ListResolvedKeyValues() = %#v, want %#v", entries, want)
	}
}

func TestConfigResolvedCommandListsResolvedKeyValuesWithoutUserFile(t *testing.T) {
	t.Setenv("HOME", "/tmp/dev-cli-home")

	initializer := &ConfigInitializer{
		configHome:   t.TempDir(),
		templateYAML: configtemplate.TemplateYAML(),
		defaultYAML:  configtemplate.DefaultYAML(),
	}

	cmd := newRootCmdWithConfigInitializer(initializer)
	var output bytes.Buffer

	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs([]string{"config", "resolved"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	want := strings.Join([]string{
		"github.api_base_url=https://api.github.com",
		"github.client_id=",
		"workspace.root=/tmp/dev-cli-home/Projects",
		"",
	}, "\n")
	if output.String() != want {
		t.Fatalf("output = %q, want %q", output.String(), want)
	}
}
