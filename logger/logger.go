package logger

import (
	"io"
	"log/slog"
	"os"
	"tg-bot-news/config"

	"github.com/natefinch/lumberjack"
)

func Setup(cfg config.Config) *slog.Logger {
	var log *slog.Logger
	var logWriter io.Writer

	if cfg.LogPath != "" {
		logWriter = &lumberjack.Logger{
			Filename:   cfg.LogPath, // Path to the log file
			MaxSize:    10,          // Max size in MB before rotation
			MaxBackups: 5,           // Max number of old log files to keep
			MaxAge:     30,          // Max number of days to retain old log files
			Compress:   true,        // Compress old log files
		}
	} else {
		logWriter = os.Stdout
	}

	switch cfg.Env {
	case "prod":
		log = slog.New(
			slog.NewTextHandler(logWriter, &slog.HandlerOptions{
				Level: slog.LevelInfo,
			}),
		)
	default:
		log = slog.New(
			slog.NewTextHandler(logWriter, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			}),
		)
	}
	return log
}
