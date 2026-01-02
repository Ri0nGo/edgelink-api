package mqtt

import (
	"context"
	"edgelink-api/internal/dataloader/receiver"
	"edgelink-api/internal/pkg/logger"
	"errors"
	"log/slog"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MQTTReceiver struct {
	mqttCfg            *MQTTConfig
	mqClient           mqtt.Client
	dropDataTopicCnt   int
	dropStatusTopicCnt int
	topicHandler       map[string]func(client mqtt.Client, msg mqtt.Message)
	dataChan           chan *receiver.Message // 设备数据队列
	statusChan         chan *receiver.Message // 设备状态队列
}

func (r *MQTTReceiver) Start(ctx context.Context) error {
	// connect mqtt broker
	if err := r.connectMQTT(); err != nil {
		return err
	}

	// wait done signal
	<-ctx.Done()

	r.Close()
	logger.Info("MQTT Receiver Close")
	return nil
}

func (r *MQTTReceiver) Name() string {
	return "MQTTReceiver"
}

func (r *MQTTReceiver) Close() error {
	if r.mqClient != nil {
		r.mqClient.Disconnect(r.mqttCfg.DisconnectTimeout)
	}
	return nil
}

func (r *MQTTReceiver) connectMQTT() error {
	mqttOpts := r.initMQTTFunc()
	r.mqClient = mqtt.NewClient(mqttOpts)
	var attempt int
	for {
		slog.Info("start to connect mqtt")
		token := r.mqClient.Connect()
		if token.WaitTimeout(r.mqttCfg.ConnectTimeout) && token.Error() == nil {
			slog.Info("MQTT connection successful", "client id", r.mqttCfg.ClientId)
			break
		}
		attempt += 1
		slog.Error("MQTT connect failed", "err", token.Error(), "attempt", attempt)
		time.Sleep(time.Minute)
		if attempt > r.mqttCfg.MaxConnectTimes {
			return errors.New("connect mqtt failed")
		}
	}
	return nil
}

func (r *MQTTReceiver) initMQTTFunc() *mqtt.ClientOptions {
	opts := mqtt.NewClientOptions().
		AddBroker(r.mqttCfg.BrokerUrl).
		SetClientID(r.mqttCfg.ClientId).
		SetUsername(r.mqttCfg.Username).
		SetPassword(r.mqttCfg.Password).
		SetCleanSession(true).
		SetKeepAlive(r.mqttCfg.KeepAlive).
		SetConnectTimeout(r.mqttCfg.ConnectTimeout).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectRetryInterval(time.Minute)

	// 当接收到与任何已知订阅不匹配的消息时将调用该MessageHandler
	//opts.SetDefaultPublishHandler(r.onMessageReceived)

	opts.OnConnect = func(client mqtt.Client) {
		for topic, handler := range r.topicHandler {
			if token := client.Subscribe(topic, r.mqttCfg.Qos, handler); token.Wait() && token.Error() != nil {
				slog.Error("subscribe topic failed", "err", token.Error(), "topic", topic)
				return
			}
		}
		logger.Info("mqtt subscribe success", "broker_url", r.mqttCfg.BrokerUrl)
	}

	opts.OnConnectionLost = func(client mqtt.Client, err error) {
		if err != nil {
			slog.Error("mqtt connect lost", "err", err, "broker_url", r.mqttCfg.BrokerUrl)
		}
	}
	return opts
}

func (r *MQTTReceiver) onMessageDataTopicHandler(client mqtt.Client, msg mqtt.Message) {
	tDetail, err := parseTopicPath(msg.Topic())
	if err != nil {
		logger.Error("parse topic failed", "err", err, "topic", msg.Topic())
		return
	}

	timer := time.NewTimer(50 * time.Millisecond)
	defer timer.Stop()

	message := &receiver.Message{
		ProductIdentifier: tDetail.productIdentifier,
		DeviceKey:         tDetail.deviceKey,
		Raw:               msg.Payload(),
		ReceivedTime:      time.Now(),
		Provider:          receiver.MQTTProvider,
	}
	select {
	case r.dataChan <- message:
		logger.Info("received data", "topic", msg.Topic(), "product_identifier", message.ProductIdentifier,
			"device_key", message.DeviceKey, "raw", string(message.Raw))
	case <-timer.C:
		r.dropDataTopicCnt += 1
		logger.Warn("data chan is full", "topic", msg.Topic(), "drop_cnt", r.dropDataTopicCnt)
	}
}

func (r *MQTTReceiver) onMessageStatusTopicHandler(client mqtt.Client, msg mqtt.Message) {
	tDetail, err := parseTopicPath(msg.Topic())
	if err != nil {
		logger.Error("parse topic failed", "err", err, "topic", msg.Topic())
		return
	}
	message := &receiver.Message{
		ProductIdentifier: tDetail.productIdentifier,
		DeviceKey:         tDetail.deviceKey,
		Raw:               msg.Payload(),
		ReceivedTime:      time.Now(),
		Provider:          receiver.MQTTProvider,
	}

	timer := time.NewTimer(50 * time.Millisecond)
	defer timer.Stop()

	select {
	case r.statusChan <- message:
		logger.Info("received data", "topic", msg.Topic(), "product_identifier", message.ProductIdentifier,
			"device_key", message.DeviceKey, "raw", string(message.Raw))
	case <-timer.C:
		r.dropStatusTopicCnt += 1
		logger.Warn("status chan is full", "topic", msg.Topic(), "drop_cnt", r.dropStatusTopicCnt)
	}
}

type topicDetail struct {
	productIdentifier string // 产品标识符
	deviceKey         string // 设备标识符
	topicLinkType     string // topic上行/下行
	topicType         string // topic类型；data / status
}

func parseTopicPath(topic string) (topicDetail, error) {
	parts := strings.Split(topic, "/")
	//   /sys/+/+/uplink/data
	if len(parts) != 6 {
		return topicDetail{}, errors.New("invalid topic len")
	}
	return topicDetail{
		productIdentifier: parts[2],
		deviceKey:         parts[3],
		topicLinkType:     parts[4],
		topicType:         parts[5],
	}, nil
}

// registerTopicHandler 订阅topic
func (r *MQTTReceiver) registerTopicHandler() {
	if r.topicHandler == nil {
		r.topicHandler = make(map[string]func(client mqtt.Client, msg mqtt.Message))
	}

	r.topicHandler[r.mqttCfg.DataTopic] = r.onMessageDataTopicHandler
	r.topicHandler[r.mqttCfg.StatusTopic] = r.onMessageStatusTopicHandler
}

func NewMQTTReceiver(cfg *MQTTConfig,
	DataChan chan *receiver.Message,
	StatusChan chan *receiver.Message,
) *MQTTReceiver {
	r := &MQTTReceiver{
		mqttCfg:    cfg,
		dataChan:   DataChan,
		statusChan: StatusChan,
	}

	r.registerTopicHandler()

	return r
}
