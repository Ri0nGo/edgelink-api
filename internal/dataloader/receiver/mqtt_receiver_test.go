package receiver

import (
	"context"
	"edgelink-api/internal/pkg/logger"
	"fmt"
	"testing"
	"time"
)

func dataChanConsumer(dataChan <-chan *Message) {
	for msg := range dataChan {
		fmt.Println("data chan msg:", msg)
	}
}

func statusChanConsumer(dataChan <-chan *Message) {
	for msg := range dataChan {
		fmt.Println("status chan msg:", msg)
	}
}

func TestNewMQTTReceiver(t *testing.T) {
	logger.InitLogger(logger.LogConfig{
		Level:         "info",
		LogFmt:        logger.LogTextFormat,
		FilePath:      "./run.log",
		ShowLogSource: true,
	})

	cfg := NewMQTTConfig(
		"tcp://127.0.0.1:1883",
		"admin",
		"123456",
		WithSSLOption(true),
	)
	ctx := context.Background()
	var dataChan = make(chan *Message, 10000)
	var statusChan = make(chan *Message, 10000)

	go dataChanConsumer(dataChan)
	go statusChanConsumer(statusChan)

	mqttReceiver := NewMQTTReceiver(ctx, cfg, dataChan, statusChan)
	mqttReceiver.Start()

	time.Sleep(time.Minute * 2)
	mqttReceiver.Close()
}

func TestNewMQTTConnect(t *testing.T) {
	logger.InitLogger(logger.LogConfig{
		Level:         "info",
		LogFmt:        logger.LogTextFormat,
		FilePath:      "./run.log",
		ShowLogSource: true,
	})

	cfg := NewMQTTConfig(
		"tcp://127.0.0.1:1883",
		"admin",
		"123456",
		WithSSLOption(true),
	)
	ctx := context.Background()
	var dataChan = make(chan *Message, 10000)
	var statusChan = make(chan *Message, 10000)

	mqttReceiver := NewMQTTReceiver(ctx, cfg, dataChan, statusChan)
	err := mqttReceiver.connectMQTT()
	if err != nil {
		t.Error(err)
		return
	}

	time.Sleep(time.Minute * 2)
	mqttReceiver.Close()
}
