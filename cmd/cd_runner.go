package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

type WorkspaceCDRunner struct {
	configInitializer *ConfigInitializer
}

func newDefaultWorkspaceCDRunner(configInitializer *ConfigInitializer) *WorkspaceCDRunner {
	if configInitializer == nil {
		configInitializer = newDefaultConfigInitializer()
	}

	return &WorkspaceCDRunner{
		configInitializer: configInitializer,
	}
}

func (r *WorkspaceCDRunner) Run(stdout io.Writer, projectName string) error {
	if stdout == nil {
		stdout = io.Discard
	}

	projectPath, err := r.Resolve(projectName)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(stdout, projectPath)
	return err
}

func (r *WorkspaceCDRunner) Resolve(projectName string) (string, error) {
	if err := r.ensureDefaults(); err != nil {
		return "", err
	}
	if err := validateWorkspaceProjectName(projectName); err != nil {
		return "", err
	}

	workspaceRoot, err := resolvedWorkspaceRoot(r.configInitializer)
	if err != nil {
		return "", err
	}

	projectPath := filepath.Join(workspaceRoot, projectName)
	info, err := os.Stat(projectPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("workspace project not found: %s", projectName)
		}

		return "", fmt.Errorf("stat workspace project %s: %w", projectName, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("workspace project is not a directory: %s", projectName)
	}

	return projectPath, nil
}

func (r *WorkspaceCDRunner) Complete(toComplete string) ([]string, cobra.ShellCompDirective, error) {
	if err := r.ensureDefaults(); err != nil {
		return nil, cobra.ShellCompDirectiveError, err
	}

	workspaceRoot, err := resolvedWorkspaceRoot(r.configInitializer)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError, err
	}

	projectNames, err := listWorkspaceProjectNamesInRoot(workspaceRoot)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError, err
	}

	completions := make([]string, 0, len(projectNames))
	for _, projectName := range projectNames {
		if strings.HasPrefix(projectName, toComplete) {
			completions = append(completions, projectName)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp, nil
}

func (r *WorkspaceCDRunner) ensureDefaults() error {
	if r == nil {
		return fmt.Errorf("workspace cd runner is not configured")
	}
	if r.configInitializer == nil {
		r.configInitializer = newDefaultConfigInitializer()
	}

	return nil
}

func validateWorkspaceProjectName(projectName string) error {
	trimmedName := strings.TrimSpace(projectName)
	if trimmedName == "" {
		return fmt.Errorf("workspace project name is empty")
	}
	if trimmedName == "." || trimmedName == ".." || filepath.Base(trimmedName) != trimmedName {
		return fmt.Errorf("workspace project name must be a single directory name")
	}

	return nil
}

func listWorkspaceProjectNamesInRoot(workspaceRoot string) ([]string, error) {
	entries, err := os.ReadDir(workspaceRoot)
	if err != nil {
		return nil, fmt.Errorf("read workspace directory: %w", err)
	}

	projectNames := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			projectNames = append(projectNames, entry.Name())
		}
	}
	sort.Strings(projectNames)

	return projectNames, nil
}
