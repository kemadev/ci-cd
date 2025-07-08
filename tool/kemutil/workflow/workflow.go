package workflow

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/kemadev/ci-cd/tool/kemutil/repotpl"
	"github.com/spf13/cobra"
)

var (
	ciImageProdURL url.URL = url.URL{
		Host: "ghcr.io",
		Path: "kemadev/ci-cd:latest",
	}
	ciImageDevURL url.URL = url.URL{
		Host: "ghcr.io",
		Path: "kemadev/ci-cd-dev:latest",
	}
)

var (
	// Hot is a flag to enable hot reload mode.
	Hot bool
	// Fix is a flag to enable fix mode.
	Fix        bool
	dockerArgs func(binary string) []string = func(binary string) []string {
		return []string{
			binary,
			"run",
			"--rm",
			"-i",
			"-t",
			"-v",
			".:/src:Z",
			"-v",
			"/tmp/gitcreds:/home/nonroot/.netrc:Z",
		}
	}
)

func getImageURL() url.URL {
	if Hot {
		slog.Debug("Hot reload mode enabled", slog.Bool("hot", Hot))
		return ciImageDevURL
	}
	slog.Debug("Hot reload mode not enabled", slog.Bool("hot", Hot))
	return ciImageProdURL
}

func prepareGitCredentials() error {
	gitToken := os.Getenv("GIT_TOKEN")
	if gitToken == "" {
		return fmt.Errorf("GIT_TOKEN environment variable is not set")
	}

	err := os.WriteFile("/tmp/gitcreds", []byte(
		fmt.Sprintf("machine %s\nlogin git\npassword %s\n",
			repotpl.RepoTemplateURL.Hostname(),
			gitToken,
		),
	), 0o600)
	if err != nil {
		return fmt.Errorf("error writing git credentials to /tmp/gitcreds: %w", err)
	}

	return nil
}

// Ci runs the CI workflows.
func Ci(cmd *cobra.Command, args []string) error {
	imageUrl := getImageURL()

	binary, err := exec.LookPath("docker")
	if err != nil {
		return fmt.Errorf("docker binary not found: %w", err)
	}

	err = prepareGitCredentials()
	if err != nil {
		return fmt.Errorf("error preparing git credentials: %w", err)
	}

	baseArgs := dockerArgs(binary)

	if Hot {
		slog.Debug("Debug mode is enabled, adding debug flag to base arguments")
		baseArgs = append(baseArgs, "-e", "RUNNER_DEBUG=1")
	}

	baseArgs = append(baseArgs, strings.TrimPrefix(imageUrl.String(), "//"))

	slog.Info("Running CI tasks")
	baseArgs = append(baseArgs, "ci")
	if Fix {
		baseArgs = append(baseArgs, "--fix")
	}
	slog.Debug("Executing CI task ci with base arguments", slog.Any("baseArgs", baseArgs))
	syscall.Exec(
		binary,
		baseArgs,
		os.Environ(),
	)

	return nil
}

// Custom runs custom commands using the CI/CD runner.
func Custom(cmd *cobra.Command, args []string) error {
	imageUrl := getImageURL()

	binary, err := exec.LookPath("docker")
	if err != nil {
		return fmt.Errorf("docker binary not found: %w", err)
	}

	err = prepareGitCredentials()
	if err != nil {
		return fmt.Errorf("error preparing git credentials: %w", err)
	}

	baseArgs := dockerArgs(binary)

	if Hot {
		slog.Debug("Debug mode is enabled, adding debug flag to base arguments")
		baseArgs = append(baseArgs, "-e", "RUNNER_DEBUG=1")
	}

	baseArgs = append(baseArgs, strings.TrimPrefix(imageUrl.String(), "//"))

	slog.Info("Running custom CI task")
	baseArgs = append(baseArgs, args...)
	if Fix {
		baseArgs = append(baseArgs, "--fix")
	}
	slog.Debug("Executing CI task custom with base arguments", slog.Any("baseArgs", baseArgs))
	syscall.Exec(
		binary,
		baseArgs,
		os.Environ(),
	)

	return nil
}
