package processor

import (
	"context"
	"edgelink-api/internal/dataloader/receiver"
	"edgelink-api/internal/dataloader/storage"
)

type Processor interface {
	Process(ctx context.Context, msg *receiver.Message, deviceID int, s storage.Storage) error
	Name() string
}

type DeviceInfo struct {
	DeviceId          int
	DeviceKey         string
	ProductIdentifier string
}
