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
	tmpDirBase string = os.TempDir()
)

var (
	// Hot is a flag to enable hot reload mode.
	Hot bool
	// Fix is a flag to enable fix mode.
	Fix        bool
	// RunnerDebug is a flag to enable debug mode for the CI/CD runner.
	RunnerDebug bool
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
			tmpDirBase + "/gitcreds:/home/nonroot/.netrc:Z",
		}
	}
	gitCredsTmpFilePath string = tmpDirBase + "/gitcreds"
)

const GitTokenEnvVarKey string = "GIT_TOKEN"

func getImageURL() url.URL {
	if Hot {
		slog.Debug("Hot reload mode enabled", slog.String("imageUrl", ciImageDevURL.String()))
		return ciImageDevURL
	}
	slog.Debug("Hot reload mode not enabled", slog.String("imageUrl", ciImageDevURL.String()))
	return ciImageProdURL
}

func prepareGitCredentials() error {
	gitToken := os.Getenv(GitTokenEnvVarKey)
	if gitToken == "" {
		return fmt.Errorf(GitTokenEnvVarKey + " environment variable is not set")
	}

	err := os.WriteFile(gitCredsTmpFilePath, []byte(
		fmt.Sprintf("machine %s\nlogin git\npassword %s\n",
			repotpl.RepoTemplateURL.Hostname(),
			gitToken,
		),
	), 0o600)
	if err != nil {
		return fmt.Errorf("error writing git credentials to "+gitCredsTmpFilePath+": %w", err)
	}

	slog.Debug("Git credentials prepared", slog.String("path", gitCredsTmpFilePath))

	return nil
}

// Ci runs the CI workflows.
func Ci(cmd *cobra.Command, args []string) error {
	slog.Debug("Running workflow CI")

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

	if RunnerDebug {
		slog.Debug("Debug mode is enabled, adding debug flag to base arguments")
		baseArgs = append(baseArgs, "-e", "RUNNER_DEBUG=1")
	}

	baseArgs = append(baseArgs, strings.TrimPrefix(imageUrl.String(), "//"))

	baseArgs = append(baseArgs, "ci")
	if Fix {
		baseArgs = append(baseArgs, "--fix")
	}

	slog.Debug("Running command", slog.Any("binary", binary), slog.Any("baseArgs", baseArgs))

	syscall.Exec(
		binary,
		baseArgs,
		os.Environ(),
	)

	return nil
}

// Custom runs custom commands using the CI/CD runner.
func Custom(cmd *cobra.Command, args []string) error {
	slog.Debug("Running workflow custom")

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

	if RunnerDebug {
		slog.Debug("Debug mode is enabled, adding debug flag to base arguments")
		baseArgs = append(baseArgs, "-e", "RUNNER_DEBUG=1")
	}

	baseArgs = append(baseArgs, strings.TrimPrefix(imageUrl.String(), "//"))

	baseArgs = append(baseArgs, args...)
	if Fix {
		slog.Debug("Fix mode is enabled, adding fix flag to base arguments")
		baseArgs = append(baseArgs, "--fix")
	}

	slog.Debug("Running command", slog.Any("binary", binary), slog.Any("baseArgs", baseArgs))

	syscall.Exec(
		binary,
		baseArgs,
		os.Environ(),
	)

	return nil
}
