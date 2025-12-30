package mqtt

import (
	"context"
	"edgelink-api/internal/dataloader/receiver"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MQTTReceiver struct {
	Ctx        context.Context
	mqttCfg    MQTTConfig
	mqClient   mqtt.Client
	dataChan   chan *receiver.Message // 设备数据队列
	statusChan chan *receiver.Message // 设备状态队列
}

func (r *MQTTReceiver) Start(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
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

func NewMQTTReceiver(ctx context.Context, cfg MQTTConfig,
	DataChan chan *receiver.Message,
	StatusChan chan *receiver.Message,
) *MQTTReceiver {
	return &MQTTReceiver{
		Ctx:        ctx,
		mqttCfg:    cfg,
		dataChan:   DataChan,
		statusChan: StatusChan,
	}
}
