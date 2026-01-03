package processor

import (
	"context"
	"edgelink-api/internal/dataloader/deviceEvent"
	"edgelink-api/internal/dataloader/receiver"
	"edgelink-api/internal/dataloader/storage"
	"edgelink-api/internal/pkg/logger"
	"encoding/json"
	"errors"
	"sync"
)

// MQTTStatusProcessor 数据处理程序
// 实现了配置通知接口，数据存储接口
type MQTTStatusProcessor struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	msgChan     chan *receiver.Message
	storage     storage.Storage
	workerCount int
	devices     map[string]*DeviceInfo
}

func (p *MQTTStatusProcessor) Notify(ctx context.Context, event *deviceEvent.Event) error {
	// receive config change notify
	// todo 后续再实现
	return nil
}

func (p *MQTTStatusProcessor) Start() error {
	// 检测 worker count 设置
	if p.workerCount < 1 {
		return errors.New("worker num must be greater than 0")
	}

	// start worker to handler message
	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go p.worker()
	}
	p.wg.Wait()
	logger.Info("close mqtt data processor")
	return nil
}

func (p *MQTTStatusProcessor) Close() {
	p.cancel()
	p.storage.Close()
}

func (p *MQTTStatusProcessor) worker() {
	defer p.wg.Done()

	for {
		select {
		case <-p.ctx.Done():
			for len(p.msgChan) > 0 {
				<-p.msgChan
			}
			return
		case msg := <-p.msgChan:
			deviceInfo, ok := p.devices[msg.DeviceKey]
			if !ok {
				logger.Error("device not found", "key", msg.DeviceKey)
				continue
			}
			var statusInfo storage.DeviceStatusInfo
			if err := json.Unmarshal(msg.Raw, &statusInfo); err != nil {
				logger.Error("failed to unmarshal device data", "key", msg.DeviceKey, "raw", msg.Raw)
				continue
			}

			err := p.storage.SaveStatus(p.ctx, deviceInfo.DeviceId, &statusInfo)
			if err != nil {
				logger.Error("save data error", "err", err)
			}
		}
	}
}

func NewMQTTStatusProcessor(ctx context.Context, msgChan chan *receiver.Message, workerCount int, storage storage.Storage) *MQTTStatusProcessor {
	pCtx, cancel := context.WithCancel(ctx)
	return &MQTTStatusProcessor{
		ctx:         pCtx,
		cancel:      cancel,
		msgChan:     msgChan,
		workerCount: workerCount,
		storage:     storage,
		devices:     make(map[string]*DeviceInfo),
	}
}
