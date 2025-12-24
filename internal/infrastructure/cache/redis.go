package cache

import (
	"context"
	"edgelink-api/internal/pkg/logger"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

func NewRedisClient(addr, auth string, db int) redis.Cmdable {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: auth, // 没有密码，默认值
		DB:       db,   // 默认DB 0
	})

	ping := client.Ping(context.Background())
	if err := ping.Err(); err != nil {
		panic(err)
	}
	logger.Info("redis connect success", "addr", addr)
	return client
}

func InitRedisCache() redis.Cmdable {
	return NewRedisClient(
		viper.GetString("redis.addr"),
		viper.GetString("redis.password"),
		viper.GetInt("redis.db"))
}
