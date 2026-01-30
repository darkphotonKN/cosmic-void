package commonhelpers

import (
	"log/slog"
	"os"
)

func SetupLogger(environment string) {
	level := slog.LevelDebug
	if environment == "production" {
		level = slog.LevelInfo
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

	slog.SetDefault(slog.New(handler))

	slog.Info("intialized logger.", "default level", level, "environment", environment)
}
