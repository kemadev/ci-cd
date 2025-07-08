package repotpl

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"os/exec"
	"syscall"

	"github.com/spf13/cobra"
)

var (
	RepoTemplateURL url.URL = url.URL{
		Scheme: "https",
		Host:   "github.com",
		Path:   "kemadev/repo-template",
	}
	copierConfigPath string = "config/copier/.copier-answers.yml"
)

// SkipAnswered indicates whether to skip questions that have already been answered during repository template update,
var SkipAnswered bool

// Init initializes the repository template.
func Init(cmd *cobra.Command, args []string) error {
	slog.Info("Initializing repository template")

	binary, err := exec.LookPath("copier")
	if err != nil {
		return fmt.Errorf("copier binary not found: %w", err)
	}

	baseArgs := []string{binary, "copy", RepoTemplateURL.String(), "."}
	slog.Debug("Running command", slog.Any("binary", binary), slog.Any("baseArgs", baseArgs))

	syscall.Exec(
		binary,
		baseArgs,
		os.Environ(),
	)

	return nil
}

// Update updates the repository template with the latest changes from upstream.
func Update(cmd *cobra.Command, args []string) error {
	slog.Info("Updating repository template")

	binary, err := exec.LookPath("copier")
	if err != nil {
		return fmt.Errorf("copier binary not found: %w", err)
	}

	baseArgs := []string{binary, "update", "--answers-file", copierConfigPath}
	if SkipAnswered {
		slog.Debug("Skip answered questions enabled", slog.Bool("skipAnswered", SkipAnswered))
		baseArgs = append(baseArgs, "--skip-answered")
	}
	slog.Debug("Running command", slog.Any("binary", binary), slog.Any("baseArgs", baseArgs))

	syscall.Exec(
		binary,
		baseArgs,
		os.Environ(),
	)

	return nil
}
