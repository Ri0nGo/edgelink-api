package notify

import (
	"context"
	"edgelink-api/internal/dataloader"
	"edgelink-api/internal/pkg/logger"
	"errors"
	"sync"
)

// ---------------- 提取公共方法部分 ---------------- //

// baseNotifier 抽离的通用部分：负责注册、分发
type baseNotifier struct {
	mu        sync.RWMutex // 保护并发注册（如果运行时可能注册）
	ctx       context.Context
	cancel    context.CancelFunc
	notifiers map[NotifyType][]NotifyHandler
}

func newBaseNotifier(ctx context.Context) *baseNotifier {
	n := &baseNotifier{
		notifiers: make(map[NotifyType][]NotifyHandler),
	}
	n.ctx, n.cancel = context.WithCancel(ctx)
	return n
}

func (b *baseNotifier) Register(notifyType NotifyType, handler NotifyHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.notifiers[notifyType] = append(b.notifiers[notifyType], handler)
}

func (b *baseNotifier) Dispatch(ctx context.Context, event *Event) {
	b.mu.RLock()
	handlers, ok := b.notifiers[event.NotifyType]
	b.mu.RUnlock()

	if !ok {
		logger.Warn("notify type not found", "type", event.NotifyType)
		return
	}

	for _, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			logger.Error("notify handler error", "err", err, "type", event.NotifyType)
			// continue 而非 break，允许其他 handler 继续执行
		}
	}
}

// ---------------- 提供包级别的方法 ---------------- //

/*
gpt 说这是一种依赖倒置，还不是很理解
本意是不想把NotifierPub放入到wire中，因为每个需要的svc都需要注入，很容易造成NewXXXSvc方法的入参过多
*/

var notifierPublisher NotifierPub = &noopPublisher{}

func InitPub(pub NotifierPub) {
	if pub != nil {
		notifierPublisher = pub
	}
}

func DeviceConfigChange(ctx context.Context, operation OperationType, data *dataloader.DeviceInfo) error {
	return notifierPublisher.DeviceConfigChange(ctx, operation, data)
}

func DevicePropChange(ctx context.Context, operation OperationType, data []*dataloader.DevicePropInfo) error {
	return notifierPublisher.DevicePropChange(ctx, operation, data)
}

type noopPublisher struct{}

func (n *noopPublisher) DeviceConfigChange(ctx context.Context, operation OperationType, data *dataloader.DeviceInfo) error {
	return errors.New("notifierPublisher is not initialized")
}

func (n *noopPublisher) DevicePropChange(ctx context.Context, operation OperationType, data []*dataloader.DevicePropInfo) error {
	return errors.New("notifierPublisher is not initialized")
}
