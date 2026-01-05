package notify

import (
	"context"
	"edgelink-api/internal/dataloader"
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
	Ts         int64         `json:"ts"`
}

// NotifyHandler 定义处理器接口
type NotifyHandler func(ctx context.Context, event *Event) error

// NotifierSub 通用的通知器订阅接口（后续 Redis、Etcd 等都实现这个接口）
type NotifierSub interface {
	// 注册处理器
	Register(notifyType NotifyType, handler NotifyHandler)
	// 手动分发事件（可选，用于测试或手动触发）

	Dispatch(ctx context.Context, event *Event)

	// 启动监听
	Start() error

	// 通知监听
	Close() error
}

// NotifierPub 发布配置接口
type NotifierPub interface {
	// 发布设备配置变更
	DeviceConfigChange(ctx context.Context, notifyType NotifyType, operation OperationType, data *dataloader.DeviceInfo) error
}
