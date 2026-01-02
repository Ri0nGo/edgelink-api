package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

/***********************
 * 1. 配置事件模型
 ***********************/
type ConfigEvent struct {
	Type      string          `json:"type"` // simple / full
	Key       string          `json:"key"`
	Version   int64           `json:"version"`
	Payload   json.RawMessage `json:"payload,omitempty"`
	Timestamp int64           `json:"ts"`
}

/***********************
 * 2. Handler 抽象
 ***********************/
type ConfigHandler interface {
	InterestedTypes() []string
	Handle(ctx context.Context, evt ConfigEvent) error
}

/***********************
 * 3. Dispatcher（核心）
 ***********************/
type ConfigDispatcher struct {
	handlers map[string][]ConfigHandler
}

func NewConfigDispatcher() *ConfigDispatcher {
	return &ConfigDispatcher{
		handlers: make(map[string][]ConfigHandler),
	}
}

func (d *ConfigDispatcher) Register(h ConfigHandler) {
	for _, t := range h.InterestedTypes() {
		d.handlers[t] = append(d.handlers[t], h)
	}
	log.Printf("registered handler: %T\n", h)
}

func (d *ConfigDispatcher) Dispatch(ctx context.Context, evt ConfigEvent) {
	hs := d.handlers[evt.Type]
	if len(hs) == 0 {
		log.Println("no handler for type:", evt.Type)
		return
	}

	for _, h := range hs {
		// 每个 handler 独立 goroutine
		go func(h ConfigHandler) {
			if err := h.Handle(ctx, evt); err != nil {
				log.Println("handle error:", err)
			}
		}(h)
	}
}

/***********************
 * 4. Redis 监听器
 ***********************/
func ListenConfigEvent(
	ctx context.Context,
	rdb *redis.Client,
	dispatcher *ConfigDispatcher,
) {
	pubsub := rdb.Subscribe(ctx, "config.event")
	defer pubsub.Close()

	// 等待订阅成功
	if _, err := pubsub.Receive(ctx); err != nil {
		log.Fatal(err)
	}

	log.Println("listening on channel: config.event")

	ch := pubsub.Channel()
	for msg := range ch {
		var evt ConfigEvent
		if err := json.Unmarshal([]byte(msg.Payload), &evt); err != nil {
			log.Println("invalid message:", err)
			continue
		}
		dispatcher.Dispatch(ctx, evt)
	}
}

/***********************
 * 5. Worker 实现
 ***********************/

// —— simpleCfg worker（1 & 3）
type SimpleWorker struct {
	Name string
}

func (w *SimpleWorker) InterestedTypes() []string {
	return []string{"simple"}
}

func (w *SimpleWorker) Handle(ctx context.Context, evt ConfigEvent) error {
	log.Printf("[%s] handle SIMPLE cfg: key=%s version=%d\n",
		w.Name, evt.Key, evt.Version)
	return nil
}

// —— fullCfg worker（2）
type FullWorker struct{}

func (w *FullWorker) InterestedTypes() []string {
	return []string{"full"}
}

func (w *FullWorker) Handle(ctx context.Context, evt ConfigEvent) error {
	log.Printf("[FULL] handle FULL cfg: key=%s version=%d\n",
		evt.Key, evt.Version)
	return nil
}

/***********************
 * 6. main
 ***********************/
func main() {
	ctx := context.Background()

	rdb := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})

	dispatcher := NewConfigDispatcher()

	// 注册 worker
	dispatcher.Register(&SimpleWorker{Name: "worker-1"})
	dispatcher.Register(&FullWorker{})
	dispatcher.Register(&SimpleWorker{Name: "worker-3"})

	// 启动监听
	go ListenConfigEvent(ctx, rdb, dispatcher)

	// 模拟发布配置变更
	go mockPublish(ctx, rdb)

	select {}
}

/***********************
 * 7. 模拟发布（demo用）
 ***********************/
func mockPublish(ctx context.Context, rdb *redis.Client) {
	time.Sleep(2 * time.Second)

	events := []ConfigEvent{
		{
			Type:      "simple",
			Key:       "device.threshold",
			Version:   1,
			Timestamp: time.Now().Unix(),
		},
		{
			Type:      "full",
			Key:       "device.full.cfg",
			Version:   2,
			Timestamp: time.Now().Unix(),
		},
	}

	for _, evt := range events {
		data, _ := json.Marshal(evt)
		rdb.Publish(ctx, "config.event", data)
		time.Sleep(2 * time.Second)
	}
}
