// Copyright 2025 kemadev
// SPDX-License-Identifier: MPL-2.0

package config

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/kemadev/ci-cd/internal/auth"
	eauth "github.com/kemadev/ci-cd/pkg/auth"
)

type Config struct {
	DebugEnabled bool
	Logger       *slog.Logger
}

const (
	DefaultConfigPath = "/var/config/"
	LocalConfigPath   = "./config/"
)

func NewConfig() (*Config, error) {
	var logLevel slog.Level

	silentEnabled := os.Getenv("RUNNER_SILENT") == "1"
	debugEnabled := os.Getenv("RUNNER_DEBUG") == "1"
	netrcEnabled := os.Getenv(eauth.NetrcEnvVarKey) != ""

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
		}

		slogFd = devNull
	} else {
		slogFd = os.Stdout
	}

	if netrcEnabled {
		err := auth.CreateNetrcFromEnv()
		if err != nil {
			return nil, fmt.Errorf("error creating netrc: %w", err)
		}
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

// Select config file, priorizing local one over default one.
func SelectFile(path string) (string, error) {
	defaultPath := DefaultConfigPath + path
	localPath := LocalConfigPath + path

	_, err := os.Stat(localPath)
	if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("error finding file: %w", err)
	} else if os.IsNotExist(err) {
		return defaultPath, nil
	}

	return localPath, nil
}
