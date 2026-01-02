package storage

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type RedisStorage struct {
	cmd redis.Cmdable
}

func (r *RedisStorage) SaveData(ctx context.Context, deviceId int, data []byte) error {
	//TODO implement me
	panic("implement me")
}

func (r *RedisStorage) SaveStatus(ctx context.Context, deviceId int, data []byte) error {
	//TODO implement me
	panic("implement me")
}

func (r *RedisStorage) Close() error {
	//TODO implement me
	panic("implement me")
}
