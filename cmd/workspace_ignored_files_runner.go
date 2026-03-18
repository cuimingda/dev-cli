package cmd

import (
	"fmt"
	"io"
	"path/filepath"
)

type WorkspaceIgnoredFileEntry struct {
	Project string
	Path    string
}

type WorkspaceIgnoredFilesRunner struct {
	configInitializer *ConfigInitializer
	runGit            workspaceGitCommandRunner
}

func newDefaultWorkspaceIgnoredFilesRunner(configInitializer *ConfigInitializer) *WorkspaceIgnoredFilesRunner {
	if configInitializer == nil {
		configInitializer = newDefaultConfigInitializer()
	}

	return &WorkspaceIgnoredFilesRunner{
		configInitializer: configInitializer,
		runGit:            runGitCommandInProject,
	}
}

func (r *WorkspaceIgnoredFilesRunner) Run(stdout io.Writer) error {
	if stdout == nil {
		stdout = io.Discard
	}

	entries, err := r.List()
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if _, err := fmt.Fprintf(stdout, "[%s] %s\n", entry.Project, entry.Path); err != nil {
			return err
		}
	}

	return nil
}

func (r *WorkspaceIgnoredFilesRunner) List() ([]WorkspaceIgnoredFileEntry, error) {
	if err := r.ensureDefaults(); err != nil {
		return nil, err
	}

	workspaceRoot, err := resolvedWorkspaceRoot(r.configInitializer)
	if err != nil {
		return nil, err
	}

	workspaces, err := listWorkspaceEntriesInRoot(workspaceRoot)
	if err != nil {
		return nil, err
	}

	ignoredEntries := []WorkspaceIgnoredFileEntry{}
	for _, workspace := range workspaces {
		projectPath := filepath.Join(workspaceRoot, workspace.LocalName)
		hasGit, err := workspaceHasGit(projectPath)
		if err != nil {
			return nil, err
		}
		if !hasGit {
			continue
		}

		ignoredPaths, err := gitIgnoredUntrackedPaths(projectPath, r.runGit)
		if err != nil {
			return nil, fmt.Errorf("list ignored files for %s: %w", workspace.LocalName, err)
		}

		for _, ignoredPath := range ignoredPaths {
			ignoredEntries = append(ignoredEntries, WorkspaceIgnoredFileEntry{
				Project: workspace.LocalName,
				Path:    ignoredPath,
			})
		}
	}

	return ignoredEntries, nil
}

func (r *WorkspaceIgnoredFilesRunner) ensureDefaults() error {
	if r == nil {
		return fmt.Errorf("workspace ignored files runner is not configured")
	}
	if r.configInitializer == nil {
		r.configInitializer = newDefaultConfigInitializer()
	}
	if r.runGit == nil {
		r.runGit = runGitCommandInProject
	}

	return nil
}
