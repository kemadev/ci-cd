package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"os/exec"
	"syscall"
)

var (
	repoTemplateURL url.URL = url.URL{
		Scheme: "https",
		Host:   "github.com",
		Path:   "kemadev/repo-template",
	}
	copierConfigPath string = "config/copier/.copier-answers.yml"
)

func runRepoTemplateTasks(args []string) error {
	slog.Debug("Running repository template tasks", slog.Any("args", args))

	var skipAnswered bool
	flagSet := flag.NewFlagSet("repo-template", flag.ExitOnError)
	flagSet.BoolVar(&skipAnswered, "skip-answered", false, "Skip answered questions in copier update")

	flagSet.Parse(args)

	if len(flagSet.Args()) != 1 {
		return fmt.Errorf(
			"expected exactly one argument for repository template tasks, got %d",
			len(flag.Args()),
		)
	}

	binary, err := exec.LookPath("copier")
	if err != nil {
		panic(err)
	}

	task := flag.Args()[0]
	switch task {
	case "init":
		slog.Info("Initializing repository template")
		syscall.Exec(
			binary,
			[]string{binary, "copy", repoTemplateURL.String(), "."},
			os.Environ(),
		)
	case "update":
		slog.Info("Updating repository template")
		syscall.Exec(
			binary,
			[]string{binary, "update", "--answers-file", copierConfigPath},
			os.Environ(),
		)
	default:
		return fmt.Errorf("unknown repository template task: %s", task)
	}
	return nil
}
