package storage

import (
	"context"
)

type Storage interface {
	SaveData(ctx context.Context, deviceId int, info *DeviceDataInfo) error
	SaveStatus(ctx context.Context, deviceId int, info *DeviceStatusInfo) error
	Close() error
}

type DeviceDataInfo struct {
	Key  string         `json:"key"`
	Ts   string         `json:"ts"`
	Data map[string]any `json:"data"` // todo 这里后面还是可以定义为float类型比较好
}

type DeviceStatusInfo struct {
	Key string `json:"key"`
	Ts  string `json:"ts"`
}
