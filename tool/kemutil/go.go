package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path"

	"github.com/kemadev/ci-cd/pkg/filesfind"
	"github.com/kemadev/ci-cd/pkg/git"
)

func runGoTasks(args []string) error {
	slog.Debug("Running Go tasks", slog.Any("args", args))

	if len(args) != 1 {
		return fmt.Errorf(
			"expected exactly one argument for Go tasks, got %d",
			len(args),
		)
	}

	task := args[0]
	switch task {
	case "update":
		slog.Info("Updating Go modules")
		mods, err := filesfind.FindFilesByExtension(filesfind.FilesFindingArgs{
			Extension: "go.mod",
			Recursive: true,
		})
		if err != nil {
			return fmt.Errorf("error finding go.mod files: %w", err)
		}
		if len(mods) == 0 {
			return fmt.Errorf("no go.mod files found in the current directory or subdirectories")
		}
		slog.Debug("Found go.mod files", slog.Any("mods", mods))
		for _, mod := range mods {
			slog.Debug("Updating Go module", slog.String("mod", mod))
			cmd := exec.Command("go", "get", "-u", "./...")
			cmd.Dir = path.Dir(mod)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("error updating Go module %s: %w", mod, err)
			}
			slog.Info("Updated Go module", slog.String("mod", mod))
		}
	case "tidy":
		slog.Info("Tidying Go modules")
		mods, err := filesfind.FindFilesByExtension(filesfind.FilesFindingArgs{
			Extension: "go.mod",
			Recursive: true,
		})
		if err != nil {
			return fmt.Errorf("error finding go.mod files: %w", err)
		}
		if len(mods) == 0 {
			return fmt.Errorf("no go.mod files found in the current directory or subdirectories")
		}
		slog.Debug("Found go.mod files", slog.Any("mods", mods))
		for _, mod := range mods {
			slog.Debug("Tidying Go module", slog.String("mod", mod))
			cmd := exec.Command("go", "mod", "tidy")
			cmd.Dir = path.Dir(mod)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("error tidying Go module %s: %w", mod, err)
			}
			slog.Info("Tidied Go module", slog.String("mod", mod))
		}
	case "init":
		slog.Info("Initializing Go module")
		basePath, err := git.GetGitBasePath()
		if err != nil {
			return fmt.Errorf("error getting git repository base path: %w", err)
		}
		if basePath == "" {
			return fmt.Errorf("error getting git repository base path")
		}

		cmd := exec.Command("go", "mod", "init", basePath)
		cmd.Dir = "."
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("error initializing Go module: %w", err)
		}
		slog.Info("Initialized Go module", slog.String("basePath", basePath))
	default:
		return fmt.Errorf("unknown Go task: %s", task)
	}

	return nil
}
