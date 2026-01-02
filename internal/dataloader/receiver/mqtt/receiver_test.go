package mqtt

import (
	"context"
	"edgelink-api/internal/dataloader/receiver"
	"edgelink-api/internal/pkg/logger"
	"fmt"
	"log/slog"
	"testing"
	"time"
)

func dataChanConsumer(dataChan <-chan *receiver.Message) {
	for msg := range dataChan {
		fmt.Println("data chan msg:", msg)
	}
}

func statusChanConsumer(dataChan <-chan *receiver.Message) {
	for msg := range dataChan {
		fmt.Println("status chan msg:", msg)
	}
}

func TestNewMQTTReceiver(t *testing.T) {
	logger.InitLogger(logger.LogConfig{
		Level:         slog.LevelInfo,
		LogFmt:        logger.LogTextFormat,
		FilePath:      "./run.log",
		ShowLogSource: true,
	})

	cfg := NewMQTTConfig(
		"tcp://127.0.0.1:1883",
		"admin",
		"123456",
	)
	ctx, cancel := context.WithCancel(context.Background())
	var dataChan = make(chan *receiver.Message, 10000)
	var statusChan = make(chan *receiver.Message, 10000)

	go dataChanConsumer(dataChan)
	go statusChanConsumer(statusChan)

	mqttReceiver := NewMQTTReceiver(cfg, dataChan, statusChan)
	go mqttReceiver.Start(ctx)

	time.Sleep(time.Minute * 2)
	cancel()
}
