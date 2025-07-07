package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"syscall"
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

func runCITasks(args []string) error {
	slog.Debug("Running CI tasks", slog.Any("args", args))

	var hot, fix bool
	flagSet := flag.NewFlagSet("ci", flag.ExitOnError)
	flagSet.BoolVar(&hot, "hot", false, "Enable hot reload mode")
	flagSet.BoolVar(&fix, "fix", false, "Enable fix mode")

	flagSet.Parse(args)

	var imageUrl url.URL

	if hot {
		slog.Debug("Hot reload mode enabled", slog.Bool("hot", hot))
		imageUrl = ciImageDevURL
	} else {
		slog.Debug("Hot reload mode not enabled", slog.Bool("hot", hot))
		imageUrl = ciImageProdURL
	}

	if fix {
		slog.Debug("Fix mode enabled", slog.Bool("fix", fix))
	}

	binary, err := exec.LookPath("docker")
	if err != nil {
		panic(err)
	}

	os.Getenv("GIT_TOKEN")
	if os.Getenv("GIT_TOKEN") == "" {
		return fmt.Errorf("GIT_TOKEN environment variable is not set")
	}

	os.WriteFile("/tmp/gitcreds", []byte(
		fmt.Sprintf("machine %s\nlogin git\npassword %s\n",
			repoTemplateURL.Hostname(),
			os.Getenv("GIT_TOKEN"),
		),
	), 0o600)

	baseArgs := []string{
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

	if debugMode {
		slog.Debug("Debug mode is enabled, adding debug flag to base arguments")
		baseArgs = append(baseArgs, "-e", "RUNNER_DEBUG=1")
	}

	baseArgs = append(baseArgs, strings.TrimPrefix(imageUrl.String(), "//"))

	task := flagSet.Arg(0)
	switch task {
	case "ci":
		slog.Info("Running CI tasks")
		baseArgs = append(baseArgs, "ci")
		if fix {
			baseArgs = append(baseArgs, "--fix")
		}
		slog.Debug("Executing CI task ci with base arguments", slog.Any("baseArgs", baseArgs))
		syscall.Exec(
			binary,
			baseArgs,
			os.Environ(),
		)
	case "custom":
		slog.Info("Running custom CI task")
		baseArgs = append(baseArgs, flagSet.Args()[1:]...)
		if fix {
			baseArgs = append(baseArgs, "--fix")
		}
		slog.Debug("Executing CI task custom with base arguments", slog.Any("baseArgs", baseArgs))
		syscall.Exec(
			binary,
			baseArgs,
			os.Environ(),
		)
	default:
		return fmt.Errorf("unknown repository template task: %s", task)
	}
	return nil
}
