package logger

import (
	"log/slog"
	"testing"
)

func TestLogger(t *testing.T) {
	err := InitLogger(LogConfig{
		slog.LevelInfo.String(),
		LogTextFormat,
		"./run.log",
		true,
	})
	if err != nil {
		t.Error(err)
	}

	log.Info("test")
	log.Debug("debug", "test", 1)
	log.Error("debug", "id", 100)
}
