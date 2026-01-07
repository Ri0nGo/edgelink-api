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

func (d *DevicePropData) TableName() string {
	return "history_data"
}

type DevicePropItem struct {
	DeviceId    int
	PropertyId  int
	PropertyKey string
}

type Persistence interface {
	// 实现该接口的函数需要处理批量查询的size
	GetDatas(ctx context.Context, deviceProps []DevicePropItem) ([]DevicePropData, error)
	// 实现该接口的函数需要处理批量插入的size
	BatchSave(ctx context.Context, datas []DevicePropData) error
}
