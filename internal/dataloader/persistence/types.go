package persistence

import (
	"context"
	"edgelink-api/internal/model"
)

type DevicePropItem struct {
	DeviceId    int
	PropertyId  int
	PropertyKey string
}

type Persistence interface {
	// 实现该接口的函数需要处理批量查询的size
	GetDatas(ctx context.Context, deviceProps []DevicePropItem) ([]model.HistoryData, error)
	// 实现该接口的函数需要处理批量插入的size
	BatchSave(ctx context.Context, datas []model.HistoryData) error
}
