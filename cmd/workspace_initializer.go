package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type WorkspaceInitializer struct {
	configInitializer *ConfigInitializer
}

func newDefaultWorkspaceInitializer(configInitializer *ConfigInitializer) *WorkspaceInitializer {
	if configInitializer == nil {
		configInitializer = newDefaultConfigInitializer()
	}

	return &WorkspaceInitializer{
		configInitializer: configInitializer,
	}
}

func (w *WorkspaceInitializer) Init() (string, error) {
	if w.configInitializer == nil {
		w.configInitializer = newDefaultConfigInitializer()
	}

	workspaceRoot, err := w.configInitializer.GetResolvedValue("workspace.root")
	if err != nil {
		return "", err
	}

	workspaceRoot = strings.TrimSpace(workspaceRoot)
	if workspaceRoot == "" {
		return "", fmt.Errorf("config value workspace.root is empty")
	}

	if _, err := os.Stat(workspaceRoot); err == nil {
		return "", fmt.Errorf("workspace directory already exists: %s", workspaceRoot)
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("stat workspace directory: %w", err)
	}

	if err := os.MkdirAll(workspaceRoot, 0o755); err != nil {
		return "", fmt.Errorf("create workspace directory: %w", err)
	}

	return workspaceRoot, nil
}
