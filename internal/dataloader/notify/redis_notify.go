package notify

import (
	"context"
	"edgelink-api/internal/pkg/logger"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisNotifier struct {
	*baseNotifier
	client      *redis.Client
	channelName string
}

func NewRedisNotifier(ctx context.Context, channelName string, client *redis.Client) *RedisNotifier {
	return &RedisNotifier{
		baseNotifier: newBaseNotifier(ctx),
		client:       client,
		channelName:  channelName,
	}
}

func (r *RedisNotifier) Start() error {
	pubsub := r.client.Subscribe(r.ctx, r.channelName)

	// 等待订阅确认（Receive 会阻塞直到确认或超时/错误）
	if _, err := pubsub.Receive(r.ctx); err != nil {
		pubsub.Close()
		return fmt.Errorf("redis subscribe failed, channel: %s, err: %v", r.channelName, err)
	}
	defer logger.Info("redis subcribe success", "channel name", r.channelName)

	go r.handlerEvent(pubsub)
	return nil
}

func (r *RedisNotifier) handlerEvent(pubsub *redis.PubSub) {
	ch := pubsub.Channel()
	defer pubsub.Close()

	for {
		select {
		case <-r.ctx.Done():
			logger.Info("redis notifier stopped")
			return
		case msg, ok := <-ch:
			if !ok {
				// channel 关闭（通常是连接问题）
				logger.Warn("redis channel closed unexpectedly")
			}

			var evt Event
			if err := json.Unmarshal([]byte(msg.Payload), &evt); err != nil {
				logger.Error("invalid event payload", "err", err, "payload", msg.Payload)
				continue
			}

			r.Dispatch(r.ctx, &evt)
		}
	}
}

// Close 通知监听
func (r *RedisNotifier) Close() error {
	if r.cancel != nil {
		r.cancel()
	}
	return nil
}
