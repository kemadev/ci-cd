package config

import (
	"fmt"
	"log/slog"
	"os"
)

type Config struct {
	DebugEnabled bool
	Logger       *slog.Logger
}

func NewConfig() (*Config, error) {
	var logLevel slog.Level

	silentEnabled := os.Getenv("RUNNER_SILENT") == "1"
	debugEnabled := os.Getenv("RUNNER_DEBUG") == "1"

	if debugEnabled {
		logLevel = slog.LevelDebug
	} else {
		logLevel = slog.LevelInfo
	}

	var slogFd *os.File
	if silentEnabled {
		devNull, err := os.Open(os.DevNull)
		if err != nil {
			return nil, fmt.Errorf("failed to open "+os.DevNull+": %w", err)
		} else {
			slogFd = devNull
		}
	} else {
		slogFd = os.Stdout
	}

	logger := slog.New(
		slog.NewTextHandler(
			slogFd,
			&slog.HandlerOptions{Level: logLevel, AddSource: debugEnabled, ReplaceAttr: nil},
		),
	)
	slog.SetDefault(logger)

	slog.Info("Start", slog.Bool("debug mode", debugEnabled))

	return &Config{
		DebugEnabled: debugEnabled,
		Logger:       logger,
	}, nil
}
