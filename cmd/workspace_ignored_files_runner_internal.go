package cmd

import (
	"bytes"
	"fmt"
	"sort"
)

func gitIgnoredUntrackedPaths(projectPath string, runGit workspaceGitCommandRunner) ([]string, error) {
	output, err := runGit(
		projectPath,
		"ls-files",
		"--others",
		"--ignored",
		"--exclude-standard",
		"--directory",
		"--no-empty-directory",
		"-z",
	)
	if err != nil {
		return nil, fmt.Errorf("run git ls-files: %w", err)
	}

	return parseGitNullSeparatedPaths(output), nil
}

func parseGitNullSeparatedPaths(output []byte) []string {
	paths := []string{}
	for _, path := range bytes.Split(output, []byte{0}) {
		if len(path) == 0 {
			continue
		}

		paths = append(paths, string(path))
	}

	sort.Strings(paths)

	return paths
}
