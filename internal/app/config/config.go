package config

import (
	"log/slog"
	"os"
)

type Config struct {
	DebugEnabled bool
	Logger       *slog.Logger
}

func NewConfig() *Config {
	var logLevel slog.Level

	silentEnabled := os.Getenv("RUNNER_SILENT") == "1"
	debugEnabled := os.Getenv("RUNNER_DEBUG") == "1"

	if silentEnabled {
		logLevel = slog.LevelError
	}
	if debugEnabled {
		logLevel = slog.LevelDebug
	} else {
		logLevel = slog.LevelInfo
	}

	logger := slog.New(
		slog.NewTextHandler(
			os.Stdout,
			&slog.HandlerOptions{Level: logLevel, AddSource: debugEnabled, ReplaceAttr: nil},
		),
	)
	slog.SetDefault(logger)

	slog.Info("Start", slog.Bool("debug mode", debugEnabled))

	return &Config{
		DebugEnabled: debugEnabled,
		Logger:       logger,
	}
}
