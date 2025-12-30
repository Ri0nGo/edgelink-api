package mqtt

import (
	"edgelink-api/internal/utils"
	"time"
)

const (
	MQTTDataTopic                  = "/sys/+/+/uplink/data"
	MQTTStatusTopic                = "/sys/+/+/uplink/status"
	DefaultMQTTConnectTimeout      = 5 * time.Second
	DefaultDisconnectTimeout  uint = 5000
	DefaultKeepAliveTime           = 30 * time.Second
)

type MQTTConfig struct {
	BrokerUrl         string
	Username          string
	Password          string
	DataTopic         string
	StatusTopic       string
	Qos               int
	KeepAlive         time.Duration // client-server 心跳检测时间
	ConnectTimeout    time.Duration // 连接时间
	DisconnectTimeout uint          // 断开连接超时，毫秒
	ClientId          string
}

func NewMQTTConfig(brokerUrl, username, password string, opts ...MQTTOption) *MQTTConfig {
	cfg := &MQTTConfig{
		BrokerUrl:         brokerUrl,
		Username:          username,
		Password:          password,
		Qos:               0,
		DataTopic:         MQTTDataTopic,
		StatusTopic:       MQTTStatusTopic,
		KeepAlive:         DefaultKeepAliveTime,
		ConnectTimeout:    DefaultMQTTConnectTimeout,
		DisconnectTimeout: DefaultDisconnectTimeout,
		ClientId:          getMqttClientId(),
	}

	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

func getMqttClientId() string {
	return "edgelink_" + utils.RandString(10)
}

type MQTTOption func(c *MQTTConfig)

func WithClientId(clientId string) MQTTOption {
	return func(c *MQTTConfig) {
		c.ClientId = clientId
	}
}

func WithDataTopic(dataTopic string) MQTTOption {
	return func(c *MQTTConfig) {
		c.DataTopic = dataTopic
	}
}

func WithStatusTopic(statusTopic string) MQTTOption {
	return func(c *MQTTConfig) {
		c.StatusTopic = statusTopic
	}
}

func WithQos(qos int) MQTTOption {
	return func(c *MQTTConfig) {
		c.Qos = qos
	}
}

func WithConnectTimeout(timeout time.Duration) MQTTOption {
	return func(c *MQTTConfig) {
		c.ConnectTimeout = timeout
	}
}

func WithDisconnectTimeout(timeout uint) MQTTOption {
	return func(c *MQTTConfig) {
		c.DisconnectTimeout = timeout
	}
}

func WithKeepAlive(time time.Duration) MQTTOption {
	return func(c *MQTTConfig) {
		c.KeepAlive = time
	}
}
