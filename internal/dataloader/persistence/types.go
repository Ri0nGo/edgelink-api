package persistence

import (
	"context"
	"time"
)

type DevicePropData struct {
	DeviceId   int       `json:"device_id"`
	PropertyId int       `json:"property_id"`
	Ts         time.Time `json:"ts"`
	Value      float64   `json:"value"`
}

type Persistence interface {
	GetDatas(ctx context.Context, deviceIdMap map[string]string /*map[deviceId]propKey*/) ([]DevicePropData, error)
	// 实现该接口的函数需要处理批量插入的size
	BatchSave(ctx context.Context, datas []DevicePropData) error
}
