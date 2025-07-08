package util

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/kemadev/ci-cd/pkg/git"
)

func GetGoModExpectedName() (string, error) {
	workdir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error getting current working directory: %w", err)
	}

	modName, err := GetGoModExpectedNameFromPath(workdir)
	if err != nil {
		return "", fmt.Errorf("error getting expected go.mod name: %w", err)
	}

	return modName, nil
}

func GetGoModExpectedNameFromPath(path string) (string, error) {
	basePath, err := git.GetGitBasePath()
	if err != nil {
		return "", fmt.Errorf("error getting git repository base path: %w", err)
	}
	if basePath == "" {
		return "", fmt.Errorf("error getting git repository base path")
	}

	slog.Debug("Git base path found", slog.String("basePath", basePath))

	repoRoot := path
	for {
		if _, err := os.Stat(filepath.Join(repoRoot, ".git")); err == nil {
			return repoRoot, nil
		}
		parent := filepath.Dir(repoRoot)
		if parent == repoRoot {
			break // reached root
		}
		repoRoot = parent
	}

	relPath, err := filepath.Rel(repoRoot, path)
	if err != nil {
		return "", fmt.Errorf("error getting relative path: %w", err)
	}
	relPath = filepath.ToSlash(relPath)

	goModName := fmt.Sprintf("%s/%s", basePath, relPath)

	return goModName, nil
}
