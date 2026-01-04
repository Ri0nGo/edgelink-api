package processor

import (
	"context"
	"edgelink-api/internal/dataloader/storage"
)

type MQTTProcessor struct {
}

func NewMQTTProcessor() *MQTTProcessor {
	return &MQTTProcessor{}
}

func (p *MQTTProcessor) HandlerData(ctx context.Context, deviceID int, data *storage.DeviceDataInfo, storage storage.Storager) error {
	return storage.SaveData(ctx, deviceID, data)
}

func (p *MQTTProcessor) HandlerStatus(ctx context.Context, deviceID int, data *storage.DeviceStatusInfo, storage storage.Storager) error {
	return storage.SaveStatus(ctx, deviceID, data)
}
