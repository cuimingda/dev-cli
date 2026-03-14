package cmd

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	configtemplate "github.com/cuimingda/dev-cli/config"
)

func TestConfigInitializerListDefaultKeyValues(t *testing.T) {
	initializer := &ConfigInitializer{
		configHome:   t.TempDir(),
		templateYAML: configtemplate.TemplateYAML(),
		defaultYAML:  configtemplate.DefaultYAML(),
	}

	entries, err := initializer.ListDefaultKeyValues()
	if err != nil {
		t.Fatalf("ListDefaultKeyValues() returned error: %v", err)
	}

	want := []string{
		"github.api_base_url=https://api.github.com",
	}
	if !reflect.DeepEqual(entries, want) {
		t.Fatalf("ListDefaultKeyValues() = %#v, want %#v", entries, want)
	}
}

func TestConfigDefaultCommandListsKeyValues(t *testing.T) {
	initializer := &ConfigInitializer{
		configHome:   t.TempDir(),
		templateYAML: configtemplate.TemplateYAML(),
		defaultYAML:  configtemplate.DefaultYAML(),
	}

	cmd := newRootCmdWithConfigInitializer(initializer)
	var output bytes.Buffer

	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs([]string{"config", "default"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	want := strings.Join([]string{
		"github.api_base_url=https://api.github.com",
		"",
	}, "\n")
	if output.String() != want {
		t.Fatalf("output = %q, want %q", output.String(), want)
	}
}
