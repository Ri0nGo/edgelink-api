package bootstrap

import (
	"context"
	"edgelink-api/internal/dataloader/notify"
	"edgelink-api/internal/ioc"
	"fmt"
	"net/http"

	"github.com/redis/go-redis/v9"
)

func Bootstrap(ctx context.Context, container *ioc.Container, srv *http.Server) {
	RunWebServer(srv)

	// 初始化发布和订阅
	client := container.RedisCmd.(*redis.Client)
	notifierSub := notify.NewRedisNotifierSub(ctx, notify.DeviceEventChannelName, client)
	notifierPub := notify.ProvideNotifierPub(container.RedisCmd)

	// 初始化数据接收器，处理器，存储器
	InitDataLoader(ctx, container.MysqlDB, container.RedisCmd, notifierSub)

	if err := notifierSub.Start(); err != nil {
		panic(fmt.Sprintf("Failed to start notifier: %v", err))
	}

	// 发送设备配置到 redis 队列
	if err := PublishDeviceConfigToRedis(ctx, container.MysqlDB, notifierPub); err != nil {
		panic(fmt.Sprintf("Failed to start notifier: %v", err))
	}
	// 发送设备属性到 redis 队列
	if err := PublishDevicePropsToRedis(ctx, container.MysqlDB, notifierPub); err != nil {
		panic(fmt.Sprintf("Failed to start notifier: %v", err))
	}
}
