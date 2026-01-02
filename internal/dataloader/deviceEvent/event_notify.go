package deviceEvent

import (
	"context"
	"edgelink-api/internal/pkg/logger"
	"encoding/json"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

type DeviceEventNotify struct {
	cmd         *redis.Client
	channelName string
	notifiers   map[NotifyType][]EventNotify
}

func NewDeviceEventNotify(channelName string) *DeviceEventNotify {
	return &DeviceEventNotify{
		channelName: channelName,
		notifiers:   make(map[NotifyType][]EventNotify),
	}
}

func (d *DeviceEventNotify) Register(notifyType NotifyType, notifier EventNotify) {
	d.notifiers[notifyType] = append(d.notifiers[notifyType], notifier)
}

func (d *DeviceEventNotify) Dispatch(ctx context.Context, event *Event) {
	notifies, ok := d.notifiers[event.NotifyType]
	if !ok {
		logger.Warn("notify type not found", "type", event.NotifyType)
		return
	}

	for _, notifier := range notifies {
		err := notifier.Notify(ctx, event)
		if err != nil {
			logger.Error("notify error", "err", err, "name", notifier.Name())
			continue
		}
	}

}

func (d *DeviceEventNotify) Start(ctx context.Context) error {
	subscribe := d.cmd.Subscribe(ctx, d.channelName)
	defer subscribe.Close()

	if _, err := subscribe.Receive(ctx); err != nil {
		return fmt.Errorf("subscribe failed, channel name:%s, err: %v", d.channelName, err)
	}

	ch := subscribe.Channel()
	for msg := range ch {
		var evt Event
		if err := json.Unmarshal([]byte(msg.Payload), &evt); err != nil {
			log.Println("invalid message:", err)
			continue
		}
		d.Dispatch(ctx, &evt)
	}
	return nil
}
