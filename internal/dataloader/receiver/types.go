package receiver

import "context"

type ProviderType string

const (
	MQTTProvider ProviderType = "mqtt"
	HTTPProvider ProviderType = "http"
)

// Receiver 接收器接口
type Receiver interface {
	// Start 开始接收数据，会阻塞直到 ctx 被取消
	// 实现方应该在收到消息后解析并发送到对应的 channel
	Start(ctx context.Context) error
	Name() string
	Close() error
}

type Message struct {
	ProductIdentifier string // 产品标识符
	DeviceKey         string // 设备标识符
	Raw               any    // 原始数据
	Provider          ProviderType
}
