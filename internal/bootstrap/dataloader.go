package bootstrap

import (
	"context"
	"edgelink-api/internal/dataloader"
	"edgelink-api/internal/dataloader/notify"
	"edgelink-api/internal/dataloader/processor"
	"edgelink-api/internal/dataloader/receiver"
	"edgelink-api/internal/dataloader/storage"
	"edgelink-api/internal/pkg/logger"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

const (
	DefaultDataChannelLength   = 10000
	DefaultStatusChannelLength = 10000
)

const (
	storageDataPrefix   = "device:"
	storageStatusPrefix = "device:status"
)

func InitDataLoader(ctx context.Context, cmd redis.Cmdable, notifier notify.NotifierSub) {
	logger.Info("init data loader")

	var dataChan = make(chan *receiver.Message, DefaultDataChannelLength)
	var statusChan = make(chan *receiver.Message, DefaultStatusChannelLength)

	// 初始化存储器
	dataStorager := initRedisStorager(cmd, "")
	statusStorager := initRedisStorager(cmd, "")

	// 初始化接收器
	mqttReceiver := initMQTTReceiver(ctx, fmt.Sprintf("tcp://%s:%d", viper.GetString("mqtt.host"), viper.GetInt("mqtt.port")),
		viper.GetString("mqtt.username"),
		viper.GetString("mqtt.password"),
		dataChan, statusChan)

	// 初始化处理器
	processors := initMqttProcessor(ctx, dataChan, statusChan, dataStorager, statusStorager, notifier)

	logger.Info("init data loader completed")

	go func() {
		<-ctx.Done()
		mqttReceiver.Close()
		for _, proc := range processors {
			proc.Close()
		}
		dataStorager.Close()
		statusStorager.Close()

		close(dataChan)
		close(statusChan)
	}()
}

func initMQTTReceiver(ctx context.Context, brokerUrl, username, password string, dataChan, statusChan chan *receiver.Message) receiver.Receiver {
	cfg := receiver.NewMQTTConfig(brokerUrl, username, password)
	mqttReceiver := receiver.NewMQTTReceiver(ctx, cfg, dataChan, statusChan)
	err := mqttReceiver.Start()
	if err != nil {
		logger.Error("mqtt receiver start failed", "err", err)
		panic(err)
	}
	return mqttReceiver
}

func initRedisStorager(cmd redis.Cmdable, prefix string) storage.Storager {
	return storage.NewRedisStorage(cmd, prefix)
}

func initMqttProcessor(ctx context.Context, dataChan,
	statusChan chan *receiver.Message,
	dataS, statusS storage.Storager,
	notifier notify.NotifierSub) []processor.ProcessorFactory {

	mqttProcessor := processor.NewMQTTProcessor()
	genericDataProcessor := processor.NewGenericProcessor(ctx, dataChan, 3, dataS)
	processor.RegisterHandler(genericDataProcessor, dataloader.MsgTypeData, mqttProcessor.HandlerData, "handler mqtt data")
	err := genericDataProcessor.Start()
	if err != nil {
		logger.Error("data processor start failed", "err", err)
		return nil
	}

	genericStatusProcessor := processor.NewGenericProcessor(ctx, statusChan, 3, statusS)
	processor.RegisterHandler(genericStatusProcessor, dataloader.MsgTypeStatus, mqttProcessor.HandlerStatus, "handler mqtt status")
	err = genericStatusProcessor.Start()
	if err != nil {
		logger.Error("status processor start failed", "err", err)
		return nil
	}

	// 注册设备的配置变动监听
	logger.Info("register device config subscribe by redis")
	notifier.Register(notify.DeviceNotifyType, genericDataProcessor.Notify)   // 设备数据处理器监听配置变动
	notifier.Register(notify.DeviceNotifyType, genericStatusProcessor.Notify) // 设备状态处理器监听配置变动

	return []processor.ProcessorFactory{genericDataProcessor, genericStatusProcessor}
}
