package processor

import (
	"context"
	"edgelink-api/internal/dataloader/receiver"
	"edgelink-api/internal/dataloader/storage"
	"edgelink-api/internal/pkg/logger"
	"errors"
	"sync"
)

type MQTTDataProcessor struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	msgChan     chan *receiver.Message
	storage     storage.Storage
	workerCount int
	devices     map[string]*DeviceInfo

	subscribeCfgName string // 订阅配置更新
}

func (p *MQTTDataProcessor) Start() error {
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
	return p.storage.Close()
}

func (p *MQTTDataProcessor) Close() {
	p.cancel()
}

func (p *MQTTDataProcessor) worker() {
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
			err := p.storage.SaveStatus(p.ctx, deviceInfo.DeviceId, msg.Raw)
			if err != nil {
				logger.Error("save data error", "err", err)
			}
		}
	}
}

func (p *MQTTDataProcessor) subscribeCfg() error {

}

func NewMQTTDataProcessor(ctx context.Context, msgChan chan *receiver.Message, workerCount int, storage storage.Storage) *MQTTDataProcessor {
	pCtx, cancel := context.WithCancel(ctx)
	return &MQTTDataProcessor{
		ctx:         pCtx,
		cancel:      cancel,
		msgChan:     msgChan,
		workerCount: workerCount,
		storage:     storage,
		devices:     make(map[string]*DeviceInfo),
	}
}
