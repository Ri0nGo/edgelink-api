package logger

import (
	"io"
	"log/slog"
	"os"
)

type LogFormat string

const (
	LogTextFormat LogFormat = "text"
	LogJsonFormat           = "json"
)

var log *slog.Logger

type LogConfig struct {
	Level         slog.Level
	LogFmt        LogFormat
	FilePath      string
	ShowLogSource bool // 打印具体的行号
}

func InitLogger(cfg LogConfig) error {
	var writer io.Writer = os.Stdout

	if cfg.FilePath != "" {
		file, err := os.OpenFile(cfg.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		writer = io.MultiWriter(os.Stdout, file)
	}

	var logOpts = &slog.HandlerOptions{
		Level: cfg.Level,
	}
	if cfg.ShowLogSource {
		logOpts.AddSource = true
	}

	var handler slog.Handler
	switch cfg.LogFmt {
	case LogJsonFormat:
		handler = slog.NewJSONHandler(writer, logOpts)
	default:
		handler = slog.NewTextHandler(writer, logOpts)
	}

	log = slog.New(handler)
	return nil
}

// ---------------- 辅助函数 ---------------- //

func parseArgs(args ...any) []slog.Attr {
	if len(args)%2 != 0 {
		return nil
	}
	attrs := make([]slog.Attr, 0, len(args)/2)
	for i := 0; i < len(args); i += 2 {
		key, ok := args[i].(string)
		if !ok {
			continue
		}
		attrs = append(attrs, slog.Any(key, args[i+1]))
	}
	return attrs
}

func Debug(msg string, args ...any) {
	log.Debug(msg, args...)
}

func Info(msg string, args ...any) {
	log.Info(msg, args...)
}

func Warn(msg string, args ...any) {
	log.Warn(msg, args...)
}

func Error(msg string, args ...any) {
	log.Error(msg, args...)
}
