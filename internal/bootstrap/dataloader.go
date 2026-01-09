package bootstrap

import (
	"context"
	"edgelink-api/internal/dataloader"
	"edgelink-api/internal/dataloader/persistence"
	"edgelink-api/internal/dataloader/processor"
	"edgelink-api/internal/dataloader/receiver"
	"edgelink-api/internal/dataloader/storage"
	"edgelink-api/internal/pkg/logger"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

const (
	DefaultDataChannelLength   = 10000
	DefaultStatusChannelLength = 10000
)

//const (
//	storageDataPrefix   = "device:"
//	storageStatusPrefix = "device:status"
//)

type DataLoaderContainer struct {
	genericPersistence     *persistence.GenericPersistence
	genericDataProcessor   *processor.GenericProcessor
	genericStatusProcessor *processor.GenericProcessor
}

func InitDataLoader(ctx context.Context, db *gorm.DB, cmd redis.Cmdable) *DataLoaderContainer {
	logger.Info("init data loader")

	var dataChan = make(chan *receiver.Message, DefaultDataChannelLength)
	var statusChan = make(chan *receiver.Message, DefaultStatusChannelLength)

	// 初始化持久化器（mysql）
	mySQLPersister := persistence.NewMySQLPersistence(cmd, db, persistence.DefaultBatchInsertSize, persistence.DefaultBatchQuerySize)
	genericPersistence := persistence.NewGenericPersistence(ctx, mySQLPersister)
	genericPersistence.Start()

	// 初始化存储器(临时存储redis)
	dataStorager := initRedisStorager(cmd)
	statusStorager := initRedisStorager(cmd)

	// 初始化接收器
	mqttReceiver := initMQTTReceiver(ctx, viper.GetString("mqtt.host"),
		viper.GetString("mqtt.username"),
		viper.GetString("mqtt.password"),
		viper.GetInt("mqtt.port"),
		viper.GetBool("mqtt.ssl"),
		dataChan, statusChan)

	// 初始化处理器
	dataProcessor, statusProcessor := initMqttProcessor(ctx, dataChan, statusChan, dataStorager, statusStorager)

	logger.Info("init data loader completed")

	go func() {
		<-ctx.Done()
		mqttReceiver.Close()
		dataProcessor.Close()
		statusProcessor.Close()
		genericPersistence.Close()
		dataStorager.Close()
		statusStorager.Close()

		close(dataChan)
		close(statusChan)
	}()

	return &DataLoaderContainer{
		genericPersistence:     genericPersistence,
		genericDataProcessor:   dataProcessor,
		genericStatusProcessor: statusProcessor,
	}
}

func initMQTTReceiver(ctx context.Context, host, username, password string, port int, ssl bool,
	dataChan, statusChan chan *receiver.Message) receiver.Receiver {
	var protocol = "tcp"
	if ssl {
		protocol = "ssl"
	}
	brokerUrl := fmt.Sprintf("%s://%s:%d", protocol, host, port)
	cfg := receiver.NewMQTTConfig(brokerUrl, username, password)
	mqttReceiver := receiver.NewMQTTReceiver(ctx, cfg, dataChan, statusChan)
	err := mqttReceiver.Start()
	if err != nil {
		logger.Error("mqtt receiver start failed", "err", err)
		panic(err)
	}
	return mqttReceiver
}

func initRedisStorager(cmd redis.Cmdable) storage.Storager {
	return storage.NewRedisStorage(cmd)
}

func initMqttProcessor(ctx context.Context, dataChan,
	statusChan chan *receiver.Message,
	dataS, statusS storage.Storager) (genericDataProcessor *processor.GenericProcessor, genericStatusProcessor *processor.GenericProcessor) {

	mqttProcessor := processor.NewMQTTProcessor()
	genericDataProcessor = processor.NewGenericProcessor(ctx, dataChan, 3, dataS)
	processor.RegisterHandler(genericDataProcessor, dataloader.MsgTypeData, mqttProcessor.HandlerData, "handler mqtt data")
	err := genericDataProcessor.Start()
	if err != nil {
		logger.Error("data processor start failed", "err", err)
		return nil, nil
	}

	genericStatusProcessor = processor.NewGenericProcessor(ctx, statusChan, 3, statusS)
	processor.RegisterHandler(genericStatusProcessor, dataloader.MsgTypeStatus, mqttProcessor.HandlerStatus, "handler mqtt status")
	err = genericStatusProcessor.Start()
	if err != nil {
		logger.Error("status processor start failed", "err", err)
		return nil, nil
	}

	return
}
