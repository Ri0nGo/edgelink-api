package processor

import (
	"context"
	"edgelink-api/internal/dataloader"
	"edgelink-api/internal/dataloader/receiver"
	"edgelink-api/internal/dataloader/storage"
	"edgelink-api/internal/pkg/logger"
	"encoding/json"
	"fmt"
	"log/slog"
	"testing"
	"time"
)

// MockStorage 模拟存储器
type MockStorage struct {
}

func (m *MockStorage) SaveData(ctx context.Context, deviceId int, info *storage.DeviceDataInfo) error {
	fmt.Println("[storage] device id: ", deviceId, " data info:", info)
	return nil
}

func (m *MockStorage) SaveStatus(ctx context.Context, deviceId int, info *storage.DeviceStatusInfo) error {
	fmt.Println("[storage] device id: ", deviceId, " status info:", info)
	return nil
}

func (m *MockStorage) Close() error {
	fmt.Println("[storage] close")
	return nil
}

// SenderMsg 发送消息
func SenderMsg(p *GenericProcessor, msg chan *receiver.Message) {
	data := map[string]float64{
		"t": 12.0,
		"h": 45,
	}
	deviceKey := "mock-device"
	dataInfo := storage.DeviceDataInfo{
		Key:  deviceKey,
		Ts:   1767494515,
		Data: data,
	}
	p.devices[deviceKey] = &dataloader.DeviceInfo{
		DeviceId:          1,
		DeviceKey:         deviceKey,
		ProductIdentifier: deviceKey,
	}
	for i := 0; i < 100; i++ {
		bytes, _ := json.Marshal(dataInfo)
		msg <- &receiver.Message{
			ProductIdentifier: deviceKey,
			DeviceKey:         deviceKey,
			MsgType:           dataloader.MsgTypeData,
			Raw:               bytes,
			Provider:          dataloader.MQTTProvider,
		}
		time.Sleep(time.Second)
	}
}

/*
sender -> msg -> processor(worker) -> storager

*/

func TestNewGenericProcessor(t *testing.T) {
	logger.InitLogger(logger.LogConfig{
		Level:         slog.LevelInfo.String(),
		LogFmt:        logger.LogTextFormat,
		FilePath:      "./run.log",
		ShowLogSource: true,
	})

	var (
		ctx   = context.Background()
		msgCh = make(chan *receiver.Message)
		ms    = MockStorage{}
	)

	processor := NewGenericProcessor(ctx, msgCh, 3, &ms)
	go SenderMsg(processor, msgCh)

	RegisterHandler(processor, dataloader.MsgTypeData, func(ctx context.Context,
		deviceID int,
		data *storage.DeviceDataInfo,
		storage storage.Storager) error {
		fmt.Println("[processor] device id: ", deviceID, " data info:", data)
		fmt.Println("start storage data")
		storage.SaveData(ctx, deviceID, data)
		return nil
	}, "mock test data info")

	processor.Start()

	time.Sleep(2 * time.Minute)
	processor.Close()

}
