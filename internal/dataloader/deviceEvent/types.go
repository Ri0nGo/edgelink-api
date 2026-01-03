package deviceEvent

import (
	"context"
	"time"
)

const DeviceEventChannelName = "device.event"

type NotifyType string

const (
	DeviceNotifyType         NotifyType = "device"
	DevicePropertyNotifyType NotifyType = "device.property"
)

type OperationType string

const (
	OperationTypeCreated OperationType = "created"
	OperationTypeUpdated OperationType = "updated"
	OperationTypeDeleted OperationType = "deleted"
)

type Event struct {
	NotifyType NotifyType    `json:"notify_type"`
	Operation  OperationType `json:"operation"`
	DeviceKey  string        `json:"device_key"`
	Payload    any           `json:"payload"` // 内容
	Ts         time.Time     `json:"ts"`
}

type EventNotify interface {
	Notify(ctx context.Context, event *Event) error
}
