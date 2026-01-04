package notify

import (
	"context"
	"edgelink-api/internal/pkg/logger"
	"sync"
)

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
