package cmd

import (
	"bytes"
	"strings"
	"testing"

	configtemplate "github.com/cuimingda/dev-cli/config"
)

func TestConfigInitializerUnsetValueRemovesExistingKey(t *testing.T) {
	initializer := &ConfigInitializer{
		configHome:   t.TempDir(),
		templateYAML: configtemplate.TemplateYAML(),
	}

	if _, err := initializer.Init(); err != nil {
		t.Fatalf("Init() returned error: %v", err)
	}

	if err := initializer.UnsetValue("github.client_id"); err != nil {
		t.Fatalf("UnsetValue() returned error: %v", err)
	}

	_, err := initializer.GetValue("github.client_id")
	if err == nil {
		t.Fatal("expected GetValue() to return an error for deleted key")
	}

	if !strings.Contains(err.Error(), "config key not found") {
		t.Fatalf("expected missing config key error, got %q", err.Error())
	}

	entries, err := initializer.ListKeyValues()
	if err != nil {
		t.Fatalf("ListKeyValues() returned error: %v", err)
	}

	if len(entries) != 0 {
		t.Fatalf("entries = %#v, want empty list", entries)
	}
}

func TestConfigInitializerUnsetValuePrunesEmptyParents(t *testing.T) {
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

	if err := initializer.UnsetValue("github.nested.callback_url"); err != nil {
		t.Fatalf("UnsetValue() returned error: %v", err)
	}

	entries, err := initializer.ListKeyValues()
	if err != nil {
		t.Fatalf("ListKeyValues() returned error: %v", err)
	}

	for _, entry := range entries {
		if strings.HasPrefix(entry, "github.nested") {
			t.Fatalf("expected nested parent to be pruned, got entry %q", entry)
		}
	}
}

func TestConfigUnsetCommandRemovesValue(t *testing.T) {
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
	cmd.SetArgs([]string{"config", "unset", "github.client_id"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	if output.String() != "" {
		t.Fatalf("expected command output to be empty on success, got %q", output.String())
	}

	_, err := initializer.GetValue("github.client_id")
	if err == nil {
		t.Fatal("expected GetValue() to return an error for deleted key")
	}
}

func TestConfigUnsetCommandReturnsErrorWhenKeyDoesNotExist(t *testing.T) {
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
	cmd.SetArgs([]string{"config", "unset", "github.unknown"})

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

func TestConfigUnsetCommandRequiresExactlyOneArg(t *testing.T) {
	initializer := &ConfigInitializer{
		configHome:   t.TempDir(),
		templateYAML: configtemplate.TemplateYAML(),
	}

	cmd := newRootCmdWithConfigInitializer(initializer)
	var output bytes.Buffer

	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs([]string{"config", "unset"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected Execute() to return an error when key argument is missing")
	}

	if !strings.Contains(err.Error(), "dev config unset requires exactly 1 argument: KEY") {
		t.Fatalf("expected argument error, got %q", err.Error())
	}

	if output.String() != "" {
		t.Fatalf("expected command output to be empty on error, got %q", output.String())
	}
}
