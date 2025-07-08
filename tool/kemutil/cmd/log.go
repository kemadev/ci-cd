package cmd

import (
	"log/slog"

	"github.com/spf13/cobra"
)

func toggleDebug(cmd *cobra.Command, args []string) {
	if cmd.Flag("debug").Value.String() == "true" {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		slog.Debug("Debug mode is enabled, setting log level to debug")
	}
}
