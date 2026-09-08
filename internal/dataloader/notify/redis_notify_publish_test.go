package notify

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"testing"

	"edgelink-api/internal/dataloader"
	"edgelink-api/internal/pkg/logger"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

// Intercept PUBLISH before any network I/O while retaining real client options.
type publishCaptureHook struct {
	args []interface{}
	err  error
}

func (h *publishCaptureHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return nil, errors.New("unexpected network access in publish test")
	}
}

func (h *publishCaptureHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		h.args = cmd.Args()
		if h.err != nil {
			return h.err
		}
		cmd.(*redis.IntCmd).SetVal(1)
		return nil
	}
}

func (h *publishCaptureHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return next
}

func TestRedisNotifierPubMetadata(t *testing.T) {
	require.NoError(t, logger.InitLogger(logger.LogConfig{Level: "info"}))
	client := redis.NewClient(&redis.Options{Addr: "unused:6379", DB: 2})
	t.Cleanup(func() { require.NoError(t, client.Close()) })
	hook := &publishCaptureHook{}
	client.AddHook(hook)
	pub := NewRedisNotifierPub(client, DeviceEventChannelName)
	err := pub.DevicePropChange(context.Background(), OperationTypeUpdated, []*dataloader.DevicePropInfo{
		{DeviceId: 4, DeviceKey: "sensor", PropertyId: 1, PropertyKey: "t"},
	})
	require.NoError(t, err)
	require.Equal(t, "publish", hook.args[0])
	require.Equal(t, "device.event", hook.args[1])
	var event Event
	require.NoError(t, json.Unmarshal(hook.args[2].([]byte), &event))
	require.NotEmpty(t, event.PublisherID)
	require.NotNil(t, event.PublisherDB)
	require.Equal(t, 2, *event.PublisherDB)
	require.Equal(t, DevicePropertyNotifyType, event.NotifyType)
	require.Equal(t, OperationTypeUpdated, event.Operation)
	require.Equal(t, "sensor", event.DeviceKey)
	props, err := json.Marshal(event.Payload)
	require.NoError(t, err)
	require.JSONEq(t, `[{"device_id":4,"device_key":"sensor","property_id":1,"property_key":"t"}]`, string(props))
}

func TestRedisNotifierPubFailure(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: "unused:6379"})
	t.Cleanup(func() { require.NoError(t, client.Close()) })
	publishErr := errors.New("publish failed")
	client.AddHook(&publishCaptureHook{err: publishErr})
	pub := NewRedisNotifierPub(client, DeviceEventChannelName)
	err := pub.DeviceConfigChange(context.Background(), OperationTypeCreated, &dataloader.DeviceInfo{DeviceId: 4})
	require.ErrorIs(t, err, publishErr)
}
