package notify

import (
	"context"
	"edgelink-api/internal/dataloader"
	"edgelink-api/internal/pkg/logger"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisNotifierSub struct {
	*baseNotifier
	client      *redis.Client
	channelName string
}

func NewRedisNotifierSub(channelName string, client *redis.Client) NotifierSub {
	return &RedisNotifierSub{
		baseNotifier: newBaseNotifier(context.Background()),
		client:       client,
		channelName:  channelName,
	}
}

func (r *RedisNotifierSub) Start() error {
	pubsub := r.client.Subscribe(r.ctx, r.channelName)

	// 等待订阅确认（Receive 会阻塞直到确认或超时/错误）
	if _, err := pubsub.Receive(r.ctx); err != nil {
		pubsub.Close()
		return fmt.Errorf("redis subscribe failed, channel: %s, err: %v", r.channelName, err)
	}

	go r.handlerEvent(pubsub)

	logger.Info("redis subcribe success", "channel", r.channelName, "redis_addr", r.client.Options().Addr, "redis_db", r.client.Options().DB)
	return nil
}

func (r *RedisNotifierSub) handlerEvent(pubsub *redis.PubSub) {
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
				return
			}

			var evt Event
			if err := json.Unmarshal([]byte(msg.Payload), &evt); err != nil {
				logger.Error("invalid event payload", "err", err, "payload", msg.Payload)
				continue
			}

			logger.Info("device config event received", "channel", r.channelName, "redis_db", r.client.Options().DB,
				"publisher_id", evt.PublisherID, "notify_type", evt.NotifyType, "operation", evt.Operation, "device_key", evt.DeviceKey, "event_ts", evt.Ts, "payload", evt.Payload)
			if evt.PublisherDB != nil && *evt.PublisherDB != r.client.Options().DB {
				logger.Warn("device config event received from different redis database", "channel", r.channelName,
					"publisher_id", evt.PublisherID, "publisher_db", *evt.PublisherDB, "redis_db", r.client.Options().DB)
			}
			r.Dispatch(r.ctx, &evt)
		}
	}
}

// Close 通知监听
func (r *RedisNotifierSub) Close() error {
	if r.cancel != nil {
		r.cancel()
	}
	return nil
}

// ---------------- 发布设备配置变更通知 ---------------- //

type RedisNotifierPub struct {
	cmd         redis.Cmdable
	channelName string
	publisherID string
}

func (r *RedisNotifierPub) DeviceConfigChange(ctx context.Context, operation OperationType,
	data *dataloader.DeviceInfo) error {
	if data == nil {
		return errors.New("data is nil")
	}

	if err := r.publishEvent(ctx, DeviceNotifyType, operation, data.DeviceKey, data); err != nil {
		return err
	}

	return nil
}

func (r *RedisNotifierPub) DevicePropChange(ctx context.Context,
	operation OperationType, data []*dataloader.DevicePropInfo) error {
	if data == nil || len(data) == 0 {
		logger.Warn("data is nil or len is zero")
		return nil
	}
	var deviceProps = make(map[string][]*dataloader.DevicePropInfo)
	for _, propInfo := range data {
		deviceProps[propInfo.DeviceKey] = append(deviceProps[propInfo.DeviceKey], propInfo)
	}
	for deviceKey, propInfos := range deviceProps {
		if err := r.publishEvent(ctx, DevicePropertyNotifyType, operation, deviceKey, propInfos); err != nil {
			return err
		}
	}

	return nil
}

func (r *RedisNotifierPub) publishEvent(ctx context.Context, notifyType NotifyType, operation OperationType,
	deviceKey string, data any) error {
	event := Event{
		PublisherID: r.publisherID,
		NotifyType:  notifyType,
		Operation:   operation,
		DeviceKey:   deviceKey,
		Payload:     data,
		Ts:          time.Now().Unix(),
	}
	if client, ok := r.cmd.(*redis.Client); ok {
		db := client.Options().DB
		event.PublisherDB = &db
	}
	bytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event failed, err: %v", err)
	}
	subscribers, err := r.cmd.Publish(ctx, r.channelName, bytes).Result()
	if err != nil {
		return fmt.Errorf("redis publish failed, channel: %s, err: %w", r.channelName, err)
	}
	logger.Info("device config event published", "channel", r.channelName, "notify_type", notifyType,
		"publisher_id", r.publisherID, "operation", operation, "device_key", deviceKey, "event_ts", event.Ts, "subscribers", subscribers, "event", string(bytes))
	if subscribers == 0 {
		logger.Warn("device config event has no subscribers", "channel", r.channelName, "notify_type", notifyType, "device_key", deviceKey)
	}
	return nil
}

func NewRedisNotifierPub(cmd redis.Cmdable, channelName string) NotifierPub {
	host, _ := os.Hostname()
	return &RedisNotifierPub{
		cmd:         cmd,
		channelName: channelName,
		publisherID: fmt.Sprintf("%s:%d", host, os.Getpid()),
	}
}

// ---------------- 初始化所有的配置通知 ---------------- //

func ProvideNotifierPub(cmd redis.Cmdable) NotifierPub {
	return NewRedisNotifierPub(cmd, DeviceEventChannelName)
}

func ProvideNotifierSub(client *redis.Client) NotifierSub {
	return NewRedisNotifierSub(DeviceEventChannelName, client)
}
