package util

import (
	"fmt"
	"os"

	"github.com/kemadev/ci-cd/pkg/git"
)

func GetGoModExpectedName() (string, error) {
	basePath, err := git.GetGitBasePath()
	if err != nil {
		return "", fmt.Errorf("error getting git repository base path: %w", err)
	}
	if basePath == "" {
		return "", fmt.Errorf("error getting git repository base path")
	}

	workdir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error getting current working directory: %w", err)
	}

	goModName := fmt.Sprintf("%s/%s", basePath, workdir)

	return goModName, nil
}

func GetGoModExpectedNameFromPath(basePath string) (string, error) {
	workdir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error getting current working directory: %w", err)
	}

	goModName := fmt.Sprintf("%s/%s", basePath, workdir)

	return goModName, nil
}
