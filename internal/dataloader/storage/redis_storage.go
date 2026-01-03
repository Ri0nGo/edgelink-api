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

func (r *RedisStorage) SaveData(ctx context.Context, deviceId int, info *DeviceDataInfo) error {
	return r.saveToRedis(ctx, deviceId, info.Data)
}

func (r *RedisStorage) SaveStatus(ctx context.Context, deviceId int, info *DeviceStatusInfo) error {
	var data = map[string]any{
		"key": info.Key,
		"ts":  info.Ts,
	}
	return r.saveToRedis(ctx, deviceId, data)
}

func (r *RedisStorage) saveToRedis(ctx context.Context, deviceId int, fields map[string]any) error {
	if r.cmd == nil {
		return errors.New("redis storage is not initialized")
	}
	return r.cmd.HSet(ctx, r.generateKey(deviceId), fields).Err()
}

func (r *RedisStorage) Close() error {
	//TODO implement me
	panic("implement me")
}

func (r *RedisStorage) generateKey(deviceId int) string {
	return fmt.Sprintf("%s:%d", r.prefix, deviceId)
}

func (r *RedisStorage) NewRedisStorage(cmd redis.Cmdable, prefix string) *RedisStorage {
	return &RedisStorage{
		cmd:    cmd,
		prefix: prefix,
	}
}
