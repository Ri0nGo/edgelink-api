package processor

import (
	"context"
	"edgelink-api/internal/dataloader"
	"edgelink-api/internal/dataloader/notify"
	"edgelink-api/internal/dataloader/receiver"
	"edgelink-api/internal/dataloader/storage"
	"edgelink-api/internal/pkg/logger"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
)

// SaveFunc 泛型化的保存函数签名
type SaveFunc[T any] func(ctx context.Context, deviceID int, data *T, storage storage.Storager) error

// HandlerConfig 每种消息类型对应的处理器配置
type HandlerConfig[T any] struct {
	Target     *T          // 用于 json.Unmarshal 的目标结构体（可以是 nil，只用类型）
	SaveMethod SaveFunc[T] // 实际调用 storage 的保存方法
	Name       string      // 用于日志等，可选
}

// GenericProcessor 统一的消息处理器（使用泛型）
type GenericProcessor struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	msgChan     chan *receiver.Message
	storager    storage.Storager
	workerCount int

	mux     sync.RWMutex
	devices map[string]*dataloader.DeviceInfo // todo 考虑后续换成 sync.Map

	// 按消息类型注册的处理器配置
	handlers map[dataloader.MsgType]Processor // 值是 *HandlerConfig[T]，T 不同
}

// NewGenericProcessor 创建处理器
func NewGenericProcessor(
	ctx context.Context,
	msgChan chan *receiver.Message,
	workerCount int,
	storage storage.Storager,
) *GenericProcessor {
	pCtx, cancel := context.WithCancel(ctx)
	return &GenericProcessor{
		ctx:         pCtx,
		cancel:      cancel,
		msgChan:     msgChan,
		storager:    storage,
		workerCount: workerCount,
		devices:     make(map[string]*dataloader.DeviceInfo),
		handlers:    make(map[dataloader.MsgType]Processor),
	}
}

// Start 开始运行
func (p *GenericProcessor) Start() error {
	if p.workerCount < 1 {
		return errors.New("worker num must be greater than 0")
	}

	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go func(name string) {
			p.worker(name)
		}(fmt.Sprintf("[worker_%d]", i))
	}

	// 通常我们不在这里 Wait()，而是让调用方决定是否阻塞
	// p.wg.Wait()
	// logger.Info("mqtt processor started")
	return nil
}

// Close 关闭
func (p *GenericProcessor) Close() {
	p.cancel()
	p.wg.Wait()
	p.storager.Close()
}

// worker 单个 worker 循环
func (p *GenericProcessor) worker(name string) {
	//logger.Info("worker started", "name", name)
	defer p.wg.Done()

	for {
		select {
		case <-p.ctx.Done():
			// 尽可能处理完剩余消息（可选行为）
			for len(p.msgChan) > 0 {
				<-p.msgChan
			}
			logger.Info("worker quit", "name", name)
			return

		case msg := <-p.msgChan:
			logger.Info("worker handler msg", "worker_name", name, "key", msg.DeviceKey)
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

			if err := handler.Process(p.ctx, msg, devInfo.DeviceId, p.storager); err != nil {
				logger.Error("process failed",
					"msg_type", handler.Name(),
					"device_key", msg.DeviceKey,
					"err", err)
			}
		}
	}
}

// getHandler 通过msg type 获取对应的 handler
func (p *GenericProcessor) getHandler(msg *receiver.Message) Processor {
	handler, ok := p.handlers[msg.MsgType]
	if !ok {
		return nil
	}
	return handler
}

// ---------------- 实现设备配置更新 ---------------- //
func (p *GenericProcessor) Notify(ctx context.Context, event *notify.Event) error {
	if event == nil {
		logger.Warn("event is nil in generic processor")
		return nil
	}
	switch event.NotifyType {
	case notify.DeviceNotifyType:
		switch event.Operation {
		case notify.OperationTypeCreated, notify.OperationTypeUpdated:
			// 期望 payload 是完整的属性 map
			if deviceInfo, ok := event.Payload.(*dataloader.DeviceInfo); ok {
				p.devices[event.DeviceKey] = deviceInfo
				logger.Info("[notify] device %s %s, full state updated\n", event.DeviceKey, event.Operation)
			} else {
				fmt.Printf("[notify] Invalid payload type for device %s %s: %T\n", event.DeviceKey, event.Operation, event.Payload)
			}

		case notify.OperationTypeDeleted:
			delete(p.devices, event.DeviceKey)
			fmt.Printf("[notify] device %s deleted\n", event.DeviceKey)
		}
	case notify.DevicePropertyNotifyType:
		logger.Info("don't handle event", "type", event.NotifyType, "operation", event.Operation)
	}
	return nil
}

// RegisterHandler 注册某种消息类型的处理器（泛型方法）
func RegisterHandler[T any](
	p *GenericProcessor,
	msgType dataloader.MsgType,
	saveFn SaveFunc[T],
	name string,
) {
	hc := &HandlerConfig[T]{
		Target:     new(T),
		SaveMethod: saveFn,
		Name:       name,
	}
	logger.Info("register handler", "type", msgType.String(), "name", hc.Name)
	p.handlers[msgType] = newGenericHandler(hc)
}

// 泛型处理器实现
type genericHandler[T any] struct {
	cfg *HandlerConfig[T]
}

func (h *genericHandler[T]) Process(ctx context.Context,
	msg *receiver.Message,
	deviceID int,
	s storage.Storager) error {
	var data T
	if err := json.Unmarshal(msg.Raw, &data); err != nil {
		return err
	}
	return h.cfg.SaveMethod(ctx, deviceID, &data, s)
}

func (h *genericHandler[T]) Name() string {
	return h.cfg.Name
}

func newGenericHandler[T any](cfg *HandlerConfig[T]) Processor {
	return &genericHandler[T]{cfg: cfg}
}
