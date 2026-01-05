package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisStorage struct {
	cmd    redis.Cmdable
	prefix string
}

// SaveData 保存设备数据，同时也会更新设备的状态数据
func (r *RedisStorage) SaveData(ctx context.Context, deviceId int, info *DeviceDataInfo) error {
	dataKey := r.generateDataKey(deviceId)
	statusKey := r.generateStatusKey(deviceId)

	dataFields := r.handleDeviceInfoData(info.Data)
	statusFields := r.handleDeviceInfoStatus(&DeviceStatusInfo{
		Key: info.Key,
		Ts:  info.Ts,
	})

	pipe := r.cmd.Pipeline()
	if len(dataFields) > 0 {
		pipe.HSet(ctx, dataKey, dataFields...)
	}
	pipe.HSet(ctx, statusKey, statusFields...)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("pipeline save device data+status failed, deviceId=%d: %w", deviceId, err)
	}
	return nil
}

func (r *RedisStorage) SaveStatus(ctx context.Context, deviceId int, info *DeviceStatusInfo) error {
	return r.saveToRedis(ctx, r.generateStatusKey(deviceId), []any{"key", info.Key, "ts", info.Ts})
}

func (r *RedisStorage) saveToRedis(ctx context.Context, key string, fields []any) error {
	if r.cmd == nil {
		return errors.New("redis storage is not initialized")
	}
	return r.cmd.HSet(ctx, key, fields...).Err()
}

func (r *RedisStorage) Close() error {
	return nil
}

func (r *RedisStorage) generateDataKey(deviceId int) string {
	return fmt.Sprintf("%s:%d:data", r.prefix, deviceId)
}

func (r *RedisStorage) generateMetaKey(deviceId int) string {
	return fmt.Sprintf("%s:%d:meta", r.prefix, deviceId)
}

func (r *RedisStorage) generateStatusKey(deviceId int) string {
	return fmt.Sprintf("%s:%d:status", r.prefix, deviceId)
}

func (r *RedisStorage) handleDeviceInfoData(data map[string]float64) []any {
	fields := make([]any, 0, len(data)*2)
	for k, v := range data {
		fields = append(fields, k, v)
	}
	return fields
}

func (r *RedisStorage) handleDeviceInfoStatus(info *DeviceStatusInfo) []any {
	return []any{"key", info.Key, "ts", info.Ts}
}

func NewRedisStorage(cmd redis.Cmdable, prefix string) *RedisStorage {
	if prefix == "" {
		prefix = "device"
	}
	return &RedisStorage{
		cmd:    cmd,
		prefix: prefix,
	}
}
