package storage

import (
	"context"
)

const DefaultRedisStoragePrefix = "device"

// DeviceDataInfo 设备数据信息
type DeviceDataInfo struct {
	Key  string             `json:"key"`
	Ts   int                `json:"ts"`
	Data map[string]float64 `json:"data"`
}

// DeviceStatusInfo 设备状态信息
type DeviceStatusInfo struct {
	Key string `json:"key"`
	Ts  int    `json:"ts"`
}

type Storager interface {
	SaveData(ctx context.Context, deviceId int, info *DeviceDataInfo) error
	SaveStatus(ctx context.Context, deviceId int, info *DeviceStatusInfo) error
	Close() error
}
