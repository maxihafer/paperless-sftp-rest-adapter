package main

import (
	"log/slog"
	"os"
)

func LogLevelFromEnv() slog.Level {
	logLevel := slog.LevelInfo

	level := os.Getenv("LOG_LEVEL")
	switch level {
	case "debug", "DEBUG":
		logLevel = slog.LevelDebug
	case "info", "INFO":
		logLevel = slog.LevelInfo
	case "warn", "WARN":
		logLevel = slog.LevelWarn
	case "error", "ERROR":
		logLevel = slog.LevelError
	}

	return logLevel
}
