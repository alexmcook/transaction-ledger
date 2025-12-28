package logger

import (
	"log/slog"
	"os"
)

func Init(isProd bool) *slog.Logger {
	var logger *slog.Logger

	if isProd {
		opts := &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}
		logger = slog.New(slog.NewJSONHandler(os.Stdout, opts))
	} else {
		opts := &slog.HandlerOptions{
			Level: slog.LevelDebug,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey {
					return slog.Attr{}
				}
				return a
			},
		}
		logger = slog.New(slog.NewTextHandler(os.Stdout, opts))
	}
	slog.SetDefault(logger)

	return logger
}
