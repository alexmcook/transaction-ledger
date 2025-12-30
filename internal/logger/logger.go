package logger

import (
	"log/slog"
	"os"
)

func Init(isProd bool) (*slog.Logger, error) {
	var logger *slog.Logger

	if isProd {
		opts := &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}
		logDir := "logs"
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, err
		}
		logFile, err := os.OpenFile(logDir+"/app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
		defer logFile.Close()
		logger = slog.New(slog.NewJSONHandler(logFile, opts))
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

	return logger, nil
}
