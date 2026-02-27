package persistence

import (
	"context"
	"edgelink-api/internal/model"
	"edgelink-api/internal/pkg/logger"
	"edgelink-api/internal/utils"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/Ri0nGo/gokit/slice"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	DefaultBatchInsertSize = 200
	DefaultBatchQuerySize  = 100
)

type MySQLPersistence struct {
	cmd redis.Cmdable
	db  *gorm.DB

	batchInsertSize int
	batchQuerySize  int
}

// GetDatas 获取设备属性数据
func (p *MySQLPersistence) GetDatas(ctx context.Context, deviceProps []DevicePropItem) ([]model.HistoryData, error) {
	devicePropGroup, err := slice.SplitChunk(deviceProps, int(math.Min(float64(p.batchQuerySize), float64(len(deviceProps)))))
	if err != nil {
		return nil, err
	}

	var results = make([]model.HistoryData, 0, len(deviceProps))
	for _, propItems := range devicePropGroup {
		data, err := p.getRedisDataByDeviceIds(ctx, propItems)
		if err != nil {
			logger.Error("get redis data failed", "err", err)
			continue
		}
		results = append(results, data...)
	}
	return results, nil
}

// BatchSave 批量插入数据
func (p *MySQLPersistence) BatchSave(ctx context.Context, datas []model.HistoryData) error {
	dataGroup, err := slice.SplitChunk(datas, int(math.Min(float64(p.batchInsertSize), float64(len(datas)))))
	if err != nil {
		return err
	}
	for _, data := range dataGroup {
		if err = p.batchInsertData(ctx, data); err != nil {
			logger.Error("batch insert data failed", "err", err)
		}
	}
	return nil
}

func (p *MySQLPersistence) getRedisDataByDeviceIds(ctx context.Context, deviceProps []DevicePropItem) ([]model.HistoryData, error) {
	if len(deviceProps) == 0 {
		return nil, nil
	}

	cmders := make([]*redis.StringCmd, len(deviceProps))
	pipeline := p.cmd.Pipeline()
	for idx, item := range deviceProps {
		key := fmt.Sprintf("device:%d:data", item.DeviceId)
		cmders[idx] = pipeline.HGet(ctx, key, item.PropertyKey)
	}
	_, err := pipeline.Exec(ctx)
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}

	currentMinuteTime := utils.GetCurrentMinuteTime().Add(-time.Minute)
	result := make([]model.HistoryData, 0, len(cmders))
	for idx, cmd := range cmders {
		valStr, err := cmd.Result()
		if err != nil {
			return nil, err
		}

		f, err := strconv.ParseFloat(valStr, 64)
		if err != nil {
			logger.Error(
				"value parse float64 failed",
				"device id", deviceProps[idx].DeviceId,
				"property id", deviceProps[idx].PropertyId,
				"value", valStr,
			)
			continue
		}
		result = append(result, model.HistoryData{
			DeviceId:   deviceProps[idx].DeviceId,
			PropertyId: deviceProps[idx].PropertyId,
			Ts:         currentMinuteTime,
			Value:      f,
		})
	}

	return result, nil
}

func (p *MySQLPersistence) batchInsertData(ctx context.Context, datas []model.HistoryData) error {
	if p.db == nil {
		return errors.New("db is nil")
	}
	return p.db.WithContext(ctx).
		Create(datas).
		Error
}

func NewMySQLPersistence(cmd redis.Cmdable, db *gorm.DB, batchInsertSize, batchQuerySize int) Persistence {
	return &MySQLPersistence{
		cmd:             cmd,
		db:              db,
		batchInsertSize: batchInsertSize,
		batchQuerySize:  batchQuerySize,
	}
}
