package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type LogFormat string

const (
	LogTextFormat LogFormat = "text"
	LogJsonFormat           = "json"
)

var log *slog.Logger

type LogConfig struct {
	Level         string // "debug;info;warning;error"
	LogFmt        LogFormat
	FilePath      string
	ShowLogSource bool // 打印具体的行号
}

func InitLogger(cfg LogConfig) error {
	var writer io.Writer = os.Stdout

	if cfg.FilePath != "" {
		if err := os.MkdirAll(filepath.Dir(cfg.FilePath), 0755); err != nil {
			return err
		}

		file, err := os.OpenFile(cfg.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		writer = io.MultiWriter(os.Stdout, file)
	}

	var logOpts = &slog.HandlerOptions{
		Level: handleLogLevel(cfg.Level),
	}
	if cfg.ShowLogSource {
		logOpts.AddSource = true
	}

	var baseHandler slog.Handler
	switch cfg.LogFmt {
	case LogJsonFormat:
		baseHandler = slog.NewJSONHandler(writer, logOpts)
	default:
		baseHandler = slog.NewTextHandler(writer, logOpts)
	}

	log = slog.New(&wrapperHandler{
		baseHandler,
		4,
	})
	return nil
}

func handleLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warning", "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

type wrapperHandler struct {
	slog.Handler
	skip int // 要往上跳过的栈帧数
}

func (h *wrapperHandler) Handle(ctx context.Context, r slog.Record) error {
	if h.skip > 0 && r.PC != 0 {
		var pcs [1]uintptr
		runtime.Callers(h.skip+1, pcs[:]) // +1 是跳过 runtime.Callers 本身
		r.PC = pcs[0]
	}
	return h.Handler.Handle(ctx, r)
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
