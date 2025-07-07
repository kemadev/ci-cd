package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
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

var debugMode bool

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
