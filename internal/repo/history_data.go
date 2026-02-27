package repo

import (
	"context"
	"edgelink-api/internal/model"
	"errors"
	"time"

	"gorm.io/gorm"
)

type IHistoryDataRepo interface {
	QueryTimeSeriesHistoryData(ctx context.Context, deviceIds, propIds []int, begin, end time.Time) ([]model.HistoryData, error)
}
type HistoryDataRepo struct {
	db *gorm.DB
}

// QueryTimeSeriesHistoryData 查询设备或属性的历史数据，设备或属性不可以都为空，时间采用半开区间
func (r *HistoryDataRepo) QueryTimeSeriesHistoryData(ctx context.Context, deviceIds, propIds []int, begin, end time.Time) ([]model.HistoryData, error) {
	if len(deviceIds) == 0 && len(propIds) == 0 {
		return nil, errors.New("设备或属性不能为空")
	}

	var historyData []model.HistoryData
	query := r.db.WithContext(ctx)

	if len(deviceIds) > 0 {
		query = query.Where("device_id IN ?", deviceIds)
	}
	if len(propIds) > 0 {
		query = query.Where("property_id IN ?", propIds)
	}

	query = query.Where("ts >= ? AND ts < ?", begin, end)
	query = query.Order("ts ASC")

	err := query.Find(&historyData).Error
	return historyData, err
}

func NewHistoryDataRepo(db *gorm.DB) IHistoryDataRepo {
	return &HistoryDataRepo{db: db}
}
