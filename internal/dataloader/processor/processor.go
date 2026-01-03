package processor

import (
	"context"
	"edgelink-api/internal/dataloader/receiver"
	"edgelink-api/internal/dataloader/storage"
	"edgelink-api/internal/pkg/logger"
	"encoding/json"
	"errors"
	"sync"
)

// 泛型化的保存函数签名
type SaveFunc[T any] func(ctx context.Context, deviceID int, data *T) error

// 每种消息类型对应的处理器配置
type HandlerConfig[T any] struct {
	Target     *T          // 用于 json.Unmarshal 的目标结构体（可以是 nil，只用类型）
	SaveMethod SaveFunc[T] // 实际调用 storage 的保存方法
	Name       string      // 用于日志等，可选
}

// 统一的消息处理器（使用泛型）
type MQTTGenericProcessor struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	msgChan     chan *receiver.Message
	storage     storage.Storage
	workerCount int
	devices     map[string]*DeviceInfo // todo 考虑后续换成 sync.Map

	// 按消息类型注册的处理器配置
	handlers map[string]any // 值是 *HandlerConfig[T]，T 不同
}

// NewMQTTGenericProcessor 创建处理器
func NewMQTTGenericProcessor(
	ctx context.Context,
	msgChan chan *receiver.Message,
	workerCount int,
	storage storage.Storage,
) *MQTTGenericProcessor {
	pCtx, cancel := context.WithCancel(ctx)
	return &MQTTGenericProcessor{
		ctx:         pCtx,
		cancel:      cancel,
		msgChan:     msgChan,
		storage:     storage,
		workerCount: workerCount,
		devices:     make(map[string]*DeviceInfo),
		handlers:    make(map[string]any),
	}
}

// RegisterHandler 注册某种消息类型的处理器（泛型方法）
func RegisterHandler[T any](
	p *MQTTGenericProcessor,
	msgType string,
	saveFn SaveFunc[T],
) {
	p.handlers[msgType] = &HandlerConfig[T]{
		Target:     new(T), // 每次 unmarshal 都会用新的实例
		SaveMethod: saveFn,
		Name:       msgType,
	}
}

// 开始运行
func (p *MQTTGenericProcessor) Start() error {
	if p.workerCount < 1 {
		return errors.New("worker num must be greater than 0")
	}

	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go p.worker()
	}

	// 通常我们不在这里 Wait()，而是让调用方决定是否阻塞
	// p.wg.Wait()
	// logger.Info("mqtt processor started")
	return nil
}

// 关闭
func (p *MQTTGenericProcessor) Close() {
	p.cancel()
	p.wg.Wait()
	p.storage.Close()
}

// 单个 worker 循环
func (p *MQTTGenericProcessor) worker() {
	defer p.wg.Done()

	for {
		select {
		case <-p.ctx.Done():
			// 尽可能处理完剩余消息（可选行为）
			for len(p.msgChan) > 0 {
				<-p.msgChan
			}
			return

		case msg := <-p.msgChan:
			devInfo, ok := p.devices[msg.DeviceKey]
			if !ok {
				logger.Error("device not found", "key", msg.DeviceKey)
				continue
			}

			handler := p.getHandler(msg)
			if handler == nil {
				logger.Warn("no handler registered for this message type", "key", msg.DeviceKey)
				continue
			}

			if err := handler.Process(p.ctx, msg, devInfo.DeviceId, p.storage); err != nil {
				logger.Error("process failed",
					"msg_type", handler.Name(),
					"device_key", msg.DeviceKey,
					"err", err)
			}
		}
	}
}

// 内部：根据消息决定使用哪个 handler（这里示例几种常见判断方式）
func (p *MQTTGenericProcessor) getHandler(msg *receiver.Message) Processor {
	handler, ok := p.handlers[msg.Type]
	if !ok {
		return nil
	}
	return handler.(Processor)
}

// 泛型处理器实现
type genericHandler[T any] struct {
	cfg *HandlerConfig[T]
}

func (h *genericHandler[T]) Process(ctx context.Context,
	msg *receiver.Message,
	deviceID int,
	s storage.Storage) error {
	var data T
	if err := json.Unmarshal(msg.Raw, &data); err != nil {
		return err
	}
	return h.cfg.SaveMethod(ctx, deviceID, &data)
}

func (h *genericHandler[T]) Name() string {
	return h.cfg.Name
}
