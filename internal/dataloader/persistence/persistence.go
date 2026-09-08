package persistence

import (
	"context"
	"edgelink-api/internal/dataloader"
	"edgelink-api/internal/dataloader/notify"
	"edgelink-api/internal/pkg/logger"
	"encoding/json"
	"sync"
	"time"
)

/*
实现了逻辑控制，具体的查询数据和插入数据由Persistence接口实现
*/

type GenericPersistence struct {
	ctx    context.Context
	cancel context.CancelFunc

	persister Persistence

	mux     sync.RWMutex
	devices map[int]map[int]*dataloader.DevicePropInfo // map[deviceId]map[propertyId]DevicePropInfo, 都是需要持久化数据
}

func NewGenericPersistence(ctx context.Context, persister Persistence) *GenericPersistence {
	newCtx, cancel := context.WithCancel(ctx)
	return &GenericPersistence{
		ctx:       newCtx,
		cancel:    cancel,
		persister: persister,
		devices:   make(map[int]map[int]*dataloader.DevicePropInfo),
	}
}

func (p *GenericPersistence) Start() error {
	go p.startPersistenceData()
	logger.Info("persistence start success")
	return nil
}

func (p *GenericPersistence) Close() {
	if p.cancel != nil {
		p.cancel()
	}
}

func (p *GenericPersistence) startPersistenceData() {
	offset := 2 * time.Second

	for {
		now := time.Now()
		next := now.Truncate(time.Minute).Add(time.Minute).Add(offset)

		select {
		case <-p.ctx.Done():
			logger.Info("persistence quit")
			return
		case <-time.After(time.Until(next)):
			p.runOnce()
		}
	}
}

func (p *GenericPersistence) runOnce() {
	started := time.Now()
	deviceProps := p.getAllDeviceProps()
	if len(deviceProps) == 0 {
		logger.Info("persistence cycle skipped", "reason", "no_configured_properties")
		return
	}
	datas, err := p.persister.GetDatas(p.ctx, deviceProps)
	if err != nil {
		logger.Error("get data failed", "property_count", len(deviceProps), "read_count", len(datas), "err", err)
	}
	if len(datas) == 0 {
		logger.Warn("persistence cycle skipped", "reason", "no_readable_data", "property_count", len(deviceProps), "elapsed_ms", time.Since(started).Milliseconds())
		return
	}

	if saveErr := p.persister.BatchSave(p.ctx, datas); saveErr != nil {
		logger.Error("batch save data failed", "property_count", len(deviceProps), "attempted_rows", len(datas), "elapsed_ms", time.Since(started).Milliseconds(), "err", saveErr)
		return
	}
	logger.Info("persistence cycle completed", "property_count", len(deviceProps), "saved_rows", len(datas),
		"skipped_properties", len(deviceProps)-len(datas), "read_failed", err != nil, "elapsed_ms", time.Since(started).Milliseconds())
}

// Notify 监听属性和设备变化事件
func (p *GenericPersistence) Notify(ctx context.Context, event *notify.Event) error {
	if event == nil {
		logger.Warn("event is nil in generic processor")
		return nil
	}
	switch event.NotifyType {
	case notify.DeviceNotifyType:
		switch event.Operation {
		case notify.OperationTypeDeleted:
			p.handleDeviceInfoDelete(event.DeviceId)
			logger.Info("notify device info deleted", "key", event.DeviceKey)
		}
	case notify.DevicePropertyNotifyType:
		logger.Info("persistence config event received", "operation", event.Operation, "device_key", event.DeviceKey, "event_ts", event.Ts)
		var propInfo []*dataloader.DevicePropInfo
		payloadBytes, err := json.Marshal(event.Payload)
		if err != nil {
			logger.Error("failed to remarshal payload", "err", err)
			return err
		}
		if err = json.Unmarshal(payloadBytes, &propInfo); err != nil {
			logger.Error("failed to unmarshal device info", "err", err, "payload", string(payloadBytes))
			return err
		}
		switch event.Operation {
		case notify.OperationTypeCreated, notify.OperationTypeUpdated:
			p.handlerDevicePropCreatedOrUpdated(propInfo)
		case notify.OperationTypeDeleted:
			p.handlerDevicePropDeleted(propInfo)
		default:
			logger.Warn("don't support operation", "type", event.NotifyType, "operation", event.Operation)
		}
	}
	return nil
}

func (p *GenericPersistence) handleDeviceInfoCreated(deviceInfo *dataloader.DeviceInfo) {
	p.mux.Lock()
	defer p.mux.Unlock()

	p.devices[deviceInfo.DeviceId] = make(map[int]*dataloader.DevicePropInfo)
}

func (p *GenericPersistence) handleDeviceInfoDelete(deviceId int) {
	p.mux.Lock()
	defer p.mux.Unlock()
	delete(p.devices, deviceId)
}

func (p *GenericPersistence) handlerDevicePropCreatedOrUpdated(deviceProps []*dataloader.DevicePropInfo) {
	p.mux.Lock()
	defer p.mux.Unlock()

	for _, prop := range deviceProps {
		if prop == nil || prop.DeviceId <= 0 || prop.PropertyId <= 0 || prop.PropertyKey == "" {
			logger.Warn("persistence config invalid", "property", prop, "action", "skip_config")
			continue
		}
		propM, ok := p.devices[prop.DeviceId]
		if !ok {
			propM = make(map[int]*dataloader.DevicePropInfo)
			p.devices[prop.DeviceId] = propM
		}
		oldKey := ""
		if old := propM[prop.PropertyId]; old != nil {
			oldKey = old.PropertyKey
		}
		copyProp := *prop
		propM[prop.PropertyId] = &copyProp
		logger.Info("persistence config property upserted", "device_id", prop.DeviceId, "device_key", prop.DeviceKey,
			"property_id", prop.PropertyId, "old_property_key", oldKey, "property_key", prop.PropertyKey)
	}
}

func (p *GenericPersistence) handlerDevicePropDeleted(deviceProps []*dataloader.DevicePropInfo) {
	p.mux.Lock()
	defer p.mux.Unlock()

	for _, prop := range deviceProps {
		if prop == nil {
			continue
		}
		if propM, ok := p.devices[prop.DeviceId]; ok {
			if old := propM[prop.PropertyId]; old != nil {
				logger.Info("persistence config property removed", "device_id", prop.DeviceId, "property_id", prop.PropertyId, "property_key", old.PropertyKey)
			}
			delete(propM, prop.PropertyId)
		}
	}
}

func (p *GenericPersistence) getAllDeviceProps() []DevicePropItem {
	p.mux.RLock()
	defer p.mux.RUnlock()

	var items []DevicePropItem
	for deviceId, propMap := range p.devices {
		for propId, info := range propMap {
			items = append(items, DevicePropItem{
				DeviceId:    deviceId,
				PropertyId:  propId,
				PropertyKey: info.PropertyKey,
			})
		}

	}
	return items
}
