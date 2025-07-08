package repotpl

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"os/exec"

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

	baseArgs := []string{"copy", RepoTemplateURL.String(), "."}
	slog.Debug("Running command", slog.Any("binary", binary), slog.Any("baseArgs", baseArgs))

	command := exec.Command(binary, baseArgs...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	err = command.Run()
	if err != nil {
		return fmt.Errorf("error running copier command: %w", err)
	}

	return nil
}

// Update updates the repository template with the latest changes from upstream.
func Update(cmd *cobra.Command, args []string) error {
	slog.Info("Updating repository template")

	binary, err := exec.LookPath("copier")
	if err != nil {
		return fmt.Errorf("copier binary not found: %w", err)
	}

	baseArgs := []string{"update", "--answers-file", copierConfigPath}
	if SkipAnswered {
		slog.Debug("Skip answered questions enabled", slog.Bool("skipAnswered", SkipAnswered))
		baseArgs = append(baseArgs, "--skip-answered")
	}
	slog.Debug("Running command", slog.Any("binary", binary), slog.Any("baseArgs", baseArgs))

	command := exec.Command(binary, baseArgs...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	err = command.Run()
	if err != nil {
		return fmt.Errorf("error running copier update command: %w", err)
	}

	return nil
}
