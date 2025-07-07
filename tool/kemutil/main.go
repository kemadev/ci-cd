package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"

	"github.com/kemadev/ci-cd/pkg/filesfind"
	kg "github.com/kemadev/ci-cd/pkg/git"
)

type cmd struct {
	usage string
	run   func(args []string) error
}

var commands = map[string]cmd{
	"ci": {
		usage: "Run CI tasks for the current repository",
		run:   runCITasks,
	},
	"go": {
		usage: "Run Go tasks for the current repository",
		run:   runGoTasks,
	},
	"repo-template": {
		usage: "Run repository template tasks",
		run:   runRepoTemplateTasks,
	},
}

var (
	debugMode       bool
	repoTemplateURL url.URL = url.URL{
		Scheme: "https",
		Host:   "github.com",
		Path:   "kemadev/repo-template",
	}
	copierConfigPath string  = "config/copier/.copier-answers.yml"
	ciImageProdURL   url.URL = url.URL{
		Host: "ghcr.io",
		Path: "kemadev/ci-cd:latest",
	}
	ciImageDevURL url.URL = url.URL{
		Host: "ghcr.io",
		Path: "kemadev/ci-cd-dev:latest",
	}
)

func init() {
	flag.BoolVar(&debugMode, "debug", false, "Enable debug mode")

	flag.Usage = usage
	flag.Parse()

	if debugMode {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		slog.Debug("Debug mode enabled", "debugMode", debugMode)
	}

	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Error: No command provided.")
		flag.Usage()
		os.Exit(1)
	}

	slog.Debug("Parsing command line arguments", slog.Any("args", flag.Args()))

	command := flag.Arg(0)
	_, exists := commands[command]
	if !exists {
		fmt.Fprintf(os.Stderr, "Error: Unknown command '%s'.\n", command)
		flag.Usage()
		os.Exit(1)
	}
}

func usage() {
	longestName := 0
	for name := range commands {
		if len(name) > longestName {
			longestName = len(name)
		}
	}
	fmt.Fprintln(os.Stderr, "Usage: "+os.Args[0]+" <command> [options]")
	fmt.Fprintln(os.Stderr, "Available commands:")
	for name, cmd := range commands {
		fmt.Printf("  %"+fmt.Sprintf("%d", longestName)+"s : %s\n", name, cmd.usage)
	}
	fmt.Fprintln(os.Stderr, "Options:")
	flag.PrintDefaults()
}

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
		basePath, err := kg.GetGitBasePath()
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

func runRepoTemplateTasks(args []string) error {
	slog.Debug("Running repository template tasks", slog.Any("args", args))
	if len(args) != 1 {
		return fmt.Errorf(
			"expected exactly one argument for repository template tasks, got %d",
			len(args),
		)
	}

	binary, err := exec.LookPath("copier")
	if err != nil {
		panic(err)
	}

	task := args[0]
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

func main() {
	command := flag.Arg(0)
	switch command {
	case "ci":
		if err := runCITasks(flag.Args()[1:]); err != nil {
			fmt.Fprintln(os.Stderr, "Error running CI tasks:", err)
			os.Exit(1)
		}
	case "go":
		if err := runGoTasks(flag.Args()[1:]); err != nil {
			fmt.Fprintln(os.Stderr, "Error running Go tasks:", err)
			os.Exit(1)
		}
	case "repo-template":
		if err := runRepoTemplateTasks(flag.Args()[1:]); err != nil {
			fmt.Fprintln(os.Stderr, "Error running repository template tasks:", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintln(os.Stderr, "Unknown command:", command)
		flag.Usage()
		os.Exit(1)
	}
}
