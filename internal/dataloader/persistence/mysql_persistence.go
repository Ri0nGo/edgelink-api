package persistence

//
//import (
//	"context"
//	"edgelink-api/internal/pkg/logger"
//	"errors"
//	"fmt"
//	"strconv"
//
//	"github.com/redis/go-redis/v9"
//	"gorm.io/gorm"
//)
//
//type MySQLPersistence struct {
//	cmd redis.Cmdable
//	db  *gorm.DB
//}
//
//func (p *MySQLPersistence) GetDatas(ctx context.Context, deviceIdMap map[string]string) ([]DevicePropData, error) {
//	deviceDataMap, err := p.getRedisDataByDeviceIds(ctx, deviceIds)
//	if err != nil {
//		logger.Error("get redis data failed", "deviceIds", ids, "err", err)
//		// 可选：是否 continue？建议根据业务决定是否中断整个任务
//		continue
//	}
//}
//
//func (p *MySQLPersistence) BatchSave(ctx context.Context, datas []DevicePropData) error {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (p *MySQLPersistence) getRedisDataByDeviceIds(ctx context.Context, ids []int) (map[int]map[string]float64, error) {
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
//
//func NewMySQLPersistence(cmd redis.Cmdable, db *gorm.DB) Persistence {
//	return &MySQLPersistence{
//		cmd: cmd,
//		db:  db,
//	}
//}
