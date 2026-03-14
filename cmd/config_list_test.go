package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	configtemplate "github.com/cuimingda/dev-cli/config"
)

func TestConfigInitializerListKeyValues(t *testing.T) {
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

	entries, err := initializer.ListKeyValues()
	if err != nil {
		t.Fatalf("ListKeyValues() returned error: %v", err)
	}

	want := []string{
		"github.api_base_url=https://api.github.com",
		"github.client_id=",
		"github.nested.callback_url=https://example.com/callback",
	}
	if !reflect.DeepEqual(entries, want) {
		t.Fatalf("ListKeyValues() = %#v, want %#v", entries, want)
	}
}

func TestConfigListCommandListsKeyValues(t *testing.T) {
	initializer := &ConfigInitializer{
		configHome:   t.TempDir(),
		templateYAML: configtemplate.TemplateYAML(),
	}

	configPath := initializer.DefaultPath()
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() returned error: %v", err)
	}

	if err := os.WriteFile(configPath, []byte(configtemplate.TemplateYAML()), 0o644); err != nil {
		t.Fatalf("WriteFile() returned error: %v", err)
	}

	cmd := newRootCmdWithConfigInitializer(initializer)
	var output bytes.Buffer

	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs([]string{"config", "list"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	want := strings.Join([]string{
		"github.client_id=",
		"",
	}, "\n")
	if output.String() != want {
		t.Fatalf("output = %q, want %q", output.String(), want)
	}
}

func TestConfigListCommandReturnsErrorWhenConfigDoesNotExist(t *testing.T) {
	initializer := &ConfigInitializer{
		configHome:   t.TempDir(),
		templateYAML: configtemplate.TemplateYAML(),
	}

	cmd := newRootCmdWithConfigInitializer(initializer)
	var output bytes.Buffer

	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs([]string{"config", "list"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected Execute() to return an error when config file does not exist")
	}

	if !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("expected missing config file error, got %q", err.Error())
	}

	if output.String() != "" {
		t.Fatalf("expected command output to be empty on error, got %q", output.String())
	}
}
