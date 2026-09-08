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
	if len(deviceProps) == 0 {
		return nil, nil
	}
	devicePropGroup, err := slice.SplitChunk(deviceProps, int(math.Min(float64(p.batchQuerySize), float64(len(deviceProps)))))
	if err != nil {
		return nil, err
	}

	var results = make([]model.HistoryData, 0, len(deviceProps))
	var readErrors []error
	for batch, propItems := range devicePropGroup {
		data, err := p.getRedisDataByDeviceIds(ctx, propItems)
		results = append(results, data...)
		if err != nil {
			logger.Error("get redis data failed", "batch", batch+1, "property_count", len(propItems), "read_count", len(data), "err", err)
			readErrors = append(readErrors, fmt.Errorf("read batch %d: %w", batch+1, err))
		}
	}
	return results, errors.Join(readErrors...)
}

// BatchSave 批量插入数据
func (p *MySQLPersistence) BatchSave(ctx context.Context, datas []model.HistoryData) error {
	if len(datas) == 0 {
		return nil
	}
	dataGroup, err := slice.SplitChunk(datas, int(math.Min(float64(p.batchInsertSize), float64(len(datas)))))
	if err != nil {
		return err
	}
	var saveErrors []error
	for batch, data := range dataGroup {
		if err = p.batchInsertData(ctx, data); err != nil {
			logger.Error("batch insert data failed", "batch", batch+1, "row_count", len(data), "sample_ts", data[0].Ts, "err", err)
			saveErrors = append(saveErrors, fmt.Errorf("insert batch %d: %w", batch+1, err))
			continue
		}
		logger.Info("persistence mysql batch saved", "batch", batch+1, "row_count", len(data), "sample_ts", data[0].Ts)
	}
	return errors.Join(saveErrors...)
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
	_, pipelineErr := pipeline.Exec(ctx)
	var readErrors []error
	if pipelineErr != nil && !errors.Is(pipelineErr, redis.Nil) {
		readErrors = append(readErrors, pipelineErr)
	}

	currentMinuteTime := utils.GetCurrentMinuteTime().Add(-time.Minute)
	result := make([]model.HistoryData, 0, len(cmders))
	for idx, cmd := range cmders {
		item := deviceProps[idx]
		attrs := []any{"device_id", item.DeviceId, "property_id", item.PropertyId,
			"property_key", item.PropertyKey, "redis_key", fmt.Sprintf("device:%d:data", item.DeviceId), "sample_ts", currentMinuteTime}
		valStr, err := cmd.Result()
		if errors.Is(err, redis.Nil) {
			logger.Warn("persistence redis field missing", append(attrs, "action", "skip_property", "err", err)...)
			continue
		}
		if err != nil {
			logger.Error("persistence redis field read failed", append(attrs, "err", err)...)
			readErrors = append(readErrors, fmt.Errorf("device %d property %d (%s): %w", item.DeviceId, item.PropertyId, item.PropertyKey, err))
			continue
		}

		f, err := strconv.ParseFloat(valStr, 64)
		if err != nil || math.IsNaN(f) || math.IsInf(f, 0) {
			logger.Warn("persistence redis value invalid", append(attrs, "value", valStr, "action", "skip_property", "err", err)...)
			continue
		}
		result = append(result, model.HistoryData{
			DeviceId:   deviceProps[idx].DeviceId,
			PropertyId: deviceProps[idx].PropertyId,
			Ts:         currentMinuteTime,
			Value:      f,
		})
	}

	return result, errors.Join(readErrors...)
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
