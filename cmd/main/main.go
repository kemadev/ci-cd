/*
ci-cd is a wrapper around CI/CD tools.

It provides consitent interface to interact with different tools and platforms, as weall as default
configuration and settings.
*/
package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/kemadev/ci-cd/internal/app/config"
	"github.com/kemadev/ci-cd/internal/app/dispatch"
)

func main() {
	startTime := time.Now()
	config, err := config.NewConfig()
	if err != nil {
		slog.Error("Failed to initialize configuration", slog.String("error", err.Error()))
		os.Exit(1)
	}

	retCode, err := dispatch.DispatchCommand(config, os.Args[1:])
	if err != nil {
		slog.Error("Error executing command", slog.String("error", err.Error()))

		retCode = 1
	}

	slog.Debug("Execution time", slog.String("duration", time.Since(startTime).String()))

	if retCode != 0 {
		os.Exit(retCode)
	}
}
