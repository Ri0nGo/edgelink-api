package persistence

//import (
//	"context"
//	"edgelink-api/internal/pkg/logger"
//	"edgelink-api/internal/utils"
//	"sync"
//	"time"
//
//	"github.com/Ri0nGo/gokit/slice"
//)

//const (
//	defaultGroupNum = 100
//)
//
//type GenericPersistence struct {
//	ctx    context.Context
//	cancel context.CancelFunc
//
//	persister Persistence
//
//	mux      sync.RWMutex
//	devices  map[int]map[int]string // map[deviceId]map[propertyId]ThingModelPropKey, 都是需要持久化数据
//	groupNum int
//}
//
//func NewGenericPersistence(ctx context.Context, persister Persistence) *GenericPersistence {
//	newCtx, cancel := context.WithCancel(ctx)
//	return &GenericPersistence{
//		ctx:       newCtx,
//		cancel:    cancel,
//		persister: persister,
//		groupNum:  defaultGroupNum,
//		devices:   make(map[int]map[int]string),
//	}
//}
//
//func (p *GenericPersistence) Start() error {
//	timer := time.NewTicker(time.Minute)
//	for {
//		select {
//		case <-p.ctx.Done():
//			logger.Info("persistence quit")
//			return nil
//		case <-timer.C:
//			err := p.handlerData()
//			if err != nil {
//				logger.Error("persistence  handle data failed", "err", err)
//			}
//		}
//	}
//}

//func (p *GenericPersistence) handlerData() error {
//	p.mux.RLock()
//	defer p.mux.RUnlock()
//	deviceIds := utils.MapKeys(p.devices)
//	if len(deviceIds) == 0 {
//		return nil
//	}
//	groups, err := slice.SplitChunk(deviceIds, p.groupNum)
//	if err != nil {
//		return err
//	}
//
//	var saveDatas []DevicePropData
//	for _, ids := range groups {
//		datas, err := p.persister.GetDatas(p.ctx, ids)
//		if err != nil {
//			logger.Error("get data failed", "ids len", len(ids), "err", err)
//			continue
//		}
//		saveDatas = append(saveDatas, datas...)
//	}
//
//	err = p.persister.BatchSave(p.ctx, saveDatas)
//	if err != nil {
//		logger.Error("batch save data failed", "count", len(saveDatas), "err", err)
//	}
//	return nil
//}

//func (p *GenericPersistence) handleData {
//	// 1. 加读锁保护整个持久化过程，避免中途 devices 被修改
//	p.mux.RLock()
//	defer p.mux.RUnlock()
//
//	deviceIds := utils.MapKeys(p.devices)
//	if len(deviceIds) == 0 {
//		return nil
//	}
//
//	// 2. 分组获取 Redis 数据（复用你之前的优化版 getRedisDataByDeviceIds）
//	groups, err := slice.SplitChunk(deviceIds, p.groupNum)
//	if err != nil {
//		return err
//	}
//
//	var saveDatas []DevicePropData
//	ts := utils.GetCurrentMinuteTime() // 统一时间戳，避免同一批数据时间略有差异
//
//	for _, ids := range groups {
//		deviceDataMap, err := p.getRedisDataByDeviceIds(p.ctx, ids)
//		if err != nil {
//			logger.Error("get redis data failed", "deviceIds", ids, "err", err)
//			// 可选：是否 continue？建议根据业务决定是否中断整个任务
//			continue
//		}
//
//		// 直接处理当前批次数据，无需合并到大 map
//		for deviceId, needProps := range p.devices {
//			redisProps, ok := deviceDataMap[deviceId]
//			if !ok || len(redisProps) == 0 {
//				continue
//			}
//
//			for propId, propKey := range needProps {
//				if val, exists := redisProps[propKey]; exists {
//					saveDatas = append(saveDatas, DevicePropData{
//						DeviceId:   deviceId,
//						PropertyId: propId,
//						Ts:         ts,
//						Value:      val,
//					})
//				}
//			}
//		}
//	}
//
//	// 3. 批量持久化
//	if len(saveDatas) == 0 {
//		return nil
//	}
//
//	if err := p.persister.BatchSave(p.ctx, saveDatas); err != nil {
//		logger.Error("batch save data failed", "count", len(saveDatas), "err", err)
//		return err // 建议返回错误，上层可决定是否重试
//	}
//
//	logger.Info("persistence success", "count", len(saveDatas))
//	return nil
//}

//func (p *GenericPersistence) getAllDeviceId() []int {
//	p.mux.RLock()
//	defer p.mux.RUnlock()
//
//	return utils.MapKeys(p.devices)
//}

//func (p *GenericPersistence) getRedisDataByDeviceIds(ctx context.Context, ids []int) (map[int]map[string]float64, error) {
//	if len(ids) == 0 {
//		return map[int]map[string]float64{}, nil
//	}
//
//	cmdMap := make(map[int]*redis.MapStringStringCmd, len(ids))
//
//	pipeline := p.cmd.Pipeline()
//
//	for _, deviceID := range ids {
//		key := fmt.Sprintf("device:%d:data", deviceID)
//		cmdMap[deviceID] = pipeline.HGetAll(ctx, key)
//	}
//
//	_, err := pipeline.Exec(ctx)
//	if err != nil && !errors.Is(err, redis.Nil) {
//		return nil, err
//	}
//
//	result := make(map[int]map[string]float64, len(cmdMap))
//
//	for deviceID, cmd := range cmdMap {
//		valMap, err := cmd.Result()
//		if err != nil {
//			return nil, err
//		}
//
//		deviceData := make(map[string]float64, len(valMap))
//		for prop, val := range valMap {
//			f, err := strconv.ParseFloat(val, 64)
//			if err != nil {
//				logger.Error(
//					"value parse float64 failed",
//					"deviceID", deviceID,
//					"property", prop,
//					"value", val,
//				)
//				continue
//			}
//			deviceData[prop] = f
//		}
//		result[deviceID] = deviceData
//	}
//
//	return result, nil
//}
