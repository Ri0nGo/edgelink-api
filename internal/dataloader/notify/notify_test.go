package notify

import (
	"context"
	"edgelink-api/internal/infrastructure/cache"
	"edgelink-api/internal/pkg/logger"
	"log/slog"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestNewDeviceEventNotify(t *testing.T) {
	logger.InitLogger(logger.LogConfig{
		Level:         slog.LevelInfo.String(),
		LogFmt:        logger.LogTextFormat,
		FilePath:      "./run.log",
		ShowLogSource: true,
	})

	cmd := cache.NewRedisClient("127.0.0.1:6379", "123456", 1)
	client, ok := cmd.(*redis.Client)
	assert.True(t, ok)

	//ctx, cancel := context.WithCancel(context.Background())
	rn := NewRedisNotifierSub(DeviceEventChannelName, client)
	rn.Register(DeviceNotifyType, func(ctx context.Context, event *Event) error {
		logger.Info("receive event", "type", event.NotifyType, "operation", event.Operation,
			"key", event.DeviceKey, "payload", event.Payload)
		return nil
	})
	rn.Start()

	<-time.After(time.Minute * 3)
	rn.Close()
}
