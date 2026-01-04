package dataloader

// ProviderType 数据供应商类型
type ProviderType string

const (
	MQTTProvider ProviderType = "mqtt"
	HTTPProvider ProviderType = "http"
)

// MsgType 消息类型
type MsgType uint8

const (
	MsgTypeData MsgType = iota + 1
	MsgTypeStatus
)

func (m MsgType) String() string {
	switch m {
	case MsgTypeStatus:
		return "status type"
	case MsgTypeData:
		return "data type"
	default:
		return "unknown type"
	}
}

// DeviceInfo 设备信息
type DeviceInfo struct {
	DeviceId          int    `json:"device_id"`
	DeviceKey         string `json:"device_key"`
	ProductIdentifier string `json:"product_identifier"`
}
