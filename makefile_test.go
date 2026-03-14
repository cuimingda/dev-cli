package main

import (
	"os"
	"strings"
	"testing"
)

func TestMakefileTargets(t *testing.T) {
	content, err := os.ReadFile("Makefile")
	if err != nil {
		t.Fatalf("read Makefile: %v", err)
	}

	checks := []struct {
		name    string
		snippet string
	}{
		{
			name:    "phony targets",
			snippet: ".PHONY: install test",
		},
		{
			name:    "install target",
			snippet: "install:\n\tgo install ./cmd/dev",
		},
		{
			name:    "test target",
			snippet: "test:\n\tgo test ./...",
		},
	}

	for _, check := range checks {
		if !strings.Contains(string(content), check.snippet) {
			t.Fatalf("Makefile missing %s snippet %q", check.name, check.snippet)
		}
	}
}
