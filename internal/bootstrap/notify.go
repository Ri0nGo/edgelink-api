package bootstrap

import (
	"context"
	"edgelink-api/internal/dataloader/notify"
	"edgelink-api/internal/pkg/logger"
	"fmt"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func InitNotify(ctx context.Context, db *gorm.DB, cmd redis.Cmdable, dataLoaderC *DataLoaderContainer) {
	client := cmd.(*redis.Client)

	notifierSub := notify.ProvideNotifierSub(client)
	notifierPub := notify.ProvideNotifierPub(cmd)

	notify.InitPub(notifierPub) // 初始化了 notify 包级别的方法（方便使用，同时保持了依赖注入且不依赖wire）

	// 注册订阅事件的 Handler
	logger.Info("register device config subscribe by redis")
	notifierSub.Register(notify.DeviceNotifyType, dataLoaderC.genericDataProcessor.Notify)       // 设备数据处理器监听配置变动
	notifierSub.Register(notify.DeviceNotifyType, dataLoaderC.genericStatusProcessor.Notify)     // 设备状态处理器监听配置变动
	notifierSub.Register(notify.DevicePropertyNotifyType, dataLoaderC.genericPersistence.Notify) // 设备属性持久化器监听配置变动

	// 开始订阅事件通知
	if err := notifierSub.Start(); err != nil {
		panic(fmt.Sprintf("Failed to start notifier: %v", err))
	}

	// 发送设备配置到 redis 队列
	if err := PublishDeviceConfigToRedis(ctx, db, notifierPub); err != nil {
		panic(fmt.Sprintf("Failed to start notifier: %v", err))
	}
	// 发送设备属性到 redis 队列
	if err := PublishDevicePropsToRedis(ctx, db, notifierPub); err != nil {
		panic(fmt.Sprintf("Failed to start notifier: %v", err))
	}

	go func() {
		<-ctx.Done()
		notifierSub.Close()
	}()
}
