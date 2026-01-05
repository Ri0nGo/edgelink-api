package storage

import (
	"context"
	"edgelink-api/internal/infrastructure/cache"
	"edgelink-api/internal/pkg/logger"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRedisStorage(t *testing.T) {
	logger.InitLogger(logger.LogConfig{
		Level:         slog.LevelInfo,
		LogFmt:        logger.LogTextFormat,
		FilePath:      "./run.log",
		ShowLogSource: true,
	})

	client := cache.NewRedisClient("127.0.0.1:6379", "123456", 1)
	s := NewRedisStorage(client, "test")
	// example data: {"key":"dht11_key_2","ts": 1767595462,"data":{"t":12.2,"h":45.90,"p_t":10.67}}
	err := s.SaveData(context.Background(), 1, &DeviceDataInfo{
		Key: "test",
		Ts:  123456,
		Data: map[string]float64{
			"test": 1.0,
		},
	})
	assert.Nil(t, err)

	// example data: {"key":"dht11_key_2","ts": 1767595462}
	err = s.SaveStatus(context.Background(), 1, &DeviceStatusInfo{
		Key: "status-key",
		Ts:  123456,
	})
	assert.Nil(t, err)
}
