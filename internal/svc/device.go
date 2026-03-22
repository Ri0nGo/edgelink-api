package svc

import (
	"context"
	"edgelink-api/internal/api/dto"
	"edgelink-api/internal/dataloader"
	"edgelink-api/internal/dataloader/notify"
	"edgelink-api/internal/model"
	bizErr "edgelink-api/internal/pkg/bizerr"
	"edgelink-api/internal/pkg/logger"
	"edgelink-api/internal/repo"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type IDeviceSvc interface {
	CreateDevice(ctx context.Context, req *dto.ReqDevice) error
	UpdateDevice(ctx context.Context, req *dto.ReqDevice) error
	DeleteDevice(ctx context.Context, id int) error
	GetDeviceById(ctx context.Context, id int) (*dto.RespDevice, error)
	GetDeviceList(ctx context.Context, search string, page dto.Page) (dto.Page, error)

	// device property
	UpdateDeviceProp(ctx context.Context, req *dto.ReqDeviceProp) error
	DeleteDeviceProps(ctx context.Context, req *dto.ReqIds) error
}

type DeviceSvc struct {
	productRepo repo.IProductRepo
	deviceRepo  repo.IDeviceRepo
	tmpRepo     repo.IThingModelPropRepo
	redisCmd    redis.Cmdable
}

func (s *DeviceSvc) CreateDevice(ctx context.Context, req *dto.ReqDevice) error {
	productModel, err := s.productRepo.GetProductById(ctx, req.ProductId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return bizErr.NewBizError("产品不存在")
		}
		return err
	}
	DeviceDao := &model.Device{
		Key:         req.Key,
		Name:        req.Name,
		ProductId:   req.ProductId,
		Address:     datatypes.NewJSONType(s.generateAddress(productModel.Identifier, req.Key)),
		Description: req.Description,
	}
	props, err := s.tmpRepo.GetThingModelPropsByModelId(ctx, productModel.ModelId)
	if err != nil {
		return err
	}
	var (
		deviceProps       = make([]model.DevicePropertyRef, len(props))
		notifyDeviceProps = make([]*dataloader.DevicePropInfo, len(props))
	)
	for i, prop := range props {
		deviceProps[i] = model.DevicePropertyRef{
			PropertyId: prop.Id,
			Persistent: true,
			StoreMode:  model.StoreModeMinute,
		}
		notifyDeviceProps[i] = &dataloader.DevicePropInfo{
			DeviceId:    DeviceDao.Id,
			DeviceKey:   DeviceDao.Key,
			PropertyId:  prop.Id,
			PropertyKey: prop.Key,
		}
	}
	if err = s.deviceRepo.CreateDevice(ctx, DeviceDao, deviceProps); err != nil {
		return err
	}

	// todo 后续可以改成发布的形式，在发布时才通知设备
	if err = notify.DeviceConfigChange(ctx, notify.OperationTypeCreated, &dataloader.DeviceInfo{
		DeviceId:          DeviceDao.Id,
		DeviceKey:         DeviceDao.Key,
		ProductIdentifier: productModel.Identifier,
	}); err != nil {
		logger.Error("notify device config failed by create", "err", err)
		return err
	}
	// 更新设备 Id
	for i, _ := range notifyDeviceProps {
		notifyDeviceProps[i].DeviceId = DeviceDao.Id
	}

	if err = notify.DevicePropChange(ctx, notify.OperationTypeCreated, notifyDeviceProps); err != nil {
		logger.Error("notify device props failed by create", "err", err)
		return err
	}
	return nil
}

func (s *DeviceSvc) generateAddress(productKey, deviceKey string) model.DeviceAddress {
	return model.DeviceAddress{
		Uplink: []model.DeviceAddressDetail{
			{
				Address: fmt.Sprintf("/sys/%s/%s/uplink/data}", productKey, deviceKey),
				Desc:    "上传设备数据至MQTT",
			}, {
				Address: fmt.Sprintf("/sys/%s/%s/uplink/status", productKey, deviceKey),
				Desc:    "上传设备状态至MQTT",
			},
		}, Downlink: []model.DeviceAddressDetail{
			{
				Address: fmt.Sprintf("/sys/%s/%s/downlink/event", productKey, deviceKey),
				Desc:    "下发事件至设备（暂未实现）",
			},
		},
	}

}

func (s *DeviceSvc) UpdateDevice(ctx context.Context, req *dto.ReqDevice) error {
	if _, err := s.productRepo.GetProductById(ctx, req.ProductId); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return bizErr.NewBizError("产品不存在")
		}
		return err
	}

	DeviceDao := &model.Device{
		Id:          req.Id,
		Name:        req.Name,
		ProductId:   req.ProductId,
		Description: req.Description,
	}
	return s.deviceRepo.UpdateDevice(ctx, DeviceDao)
}

func (s *DeviceSvc) DeleteDevice(ctx context.Context, id int) error {
	err := s.deviceRepo.DeleteDevice(ctx, id)
	if err != nil {
		return err
	}

	if err = s.deviceRepo.DeleteDevicePropByDeviceId(ctx, id); err != nil {
		return err
	}
	// todo 这里还需要删除对应的历史数据
	return nil
}

func (s *DeviceSvc) GetDeviceById(ctx context.Context, id int) (*dto.RespDevice, error) {
	DeviceDao, err := s.deviceRepo.GetDeviceById(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, bizErr.NewBizError("设备不存在")
		}
		return nil, err
	}
	productModel, err := s.productRepo.GetProductById(ctx, DeviceDao.ProductId)
	if err != nil {
		return nil, err
	}
	DeviceDao.ProductName = productModel.Name
	devices := []model.Device{DeviceDao}
	if err = s.fillDeviceStatus(ctx, devices); err != nil {
		return nil, err
	}
	DeviceDao = devices[0]

	// 查询该设备使用了哪些模型属性（设备下不允许编辑模型属性的标识符，数据类型）
	props, err := s.deviceRepo.GetDevicePropByDeviceId(ctx, id)
	if err != nil {
		return nil, err
	}
	if err = s.fillDevicePropValues(ctx, DeviceDao.Id, props); err != nil {
		return nil, err
	}

	result := &dto.RespDevice{
		Device: DeviceDao,
		Props:  props,
	}

	return result, nil
}

func (s *DeviceSvc) GetDeviceList(ctx context.Context, search string, page dto.Page) (dto.Page, error) {
	dataPage, err := s.deviceRepo.GetDeviceList(ctx, search, page)
	if err != nil {
		return dto.Page{}, err
	}
	Devices, ok := dataPage.Data.([]model.Device)
	if !ok {
		return dto.Page{}, err
	}

	// 查询产品所使用的模型
	var productIds []int
	for _, Device := range Devices {
		productIds = append(productIds, Device.ProductId)
	}

	modelsMap, err := s.getProductsMap(ctx, productIds)
	if err != nil {
		return dto.Page{}, err
	}
	for i, Device := range Devices {
		if m, ok := modelsMap[Device.ProductId]; ok {
			Devices[i].ProductName = m.Name
		}
	}

	if err = s.fillDeviceStatus(ctx, Devices); err != nil {
		return dto.Page{}, err
	}
	dataPage.Data = Devices
	return dataPage, nil
}

func (s *DeviceSvc) fillDeviceStatus(ctx context.Context, devices []model.Device) error {
	if len(devices) == 0 {
		return nil
	}
	if s.redisCmd == nil {
		return nil
	}

	pipe := s.redisCmd.Pipeline()
	cmders := make([]*redis.MapStringStringCmd, len(devices))
	for i, device := range devices {
		cmders[i] = pipe.HGetAll(ctx, fmt.Sprintf("device:%d:status", device.Id))
	}

	_, err := pipe.Exec(ctx)
	if err != nil && !errors.Is(err, redis.Nil) {
		return err
	}

	for i, cmd := range cmders {
		devices[i].Status = "unknown"
		devices[i].StatusUpdatedTime = nil

		statusData, cmdErr := cmd.Result()
		if cmdErr != nil && !errors.Is(cmdErr, redis.Nil) {
			return cmdErr
		}
		if len(statusData) == 0 {
			continue
		}

		statusKey, ok := statusData["key"]
		if !ok || statusKey == "" {
			continue
		}

		if ts, ok := statusData["ts"]; ok {
			parsedTs, parseErr := strconv.ParseInt(ts, 10, 64)
			if parseErr != nil {
				logger.Warn("parse device status ts failed", "device_id", devices[i].Id, "ts", ts, "err", parseErr)
				continue
			}

			nowTs := time.Now().Unix()
			if nowTs-parsedTs <= 60 {
				devices[i].Status = "online"
			} else {
				devices[i].Status = "offline"
			}
			devices[i].StatusUpdatedTime = &parsedTs
		}
	}

	return nil
}

func (s *DeviceSvc) fillDevicePropValues(ctx context.Context, deviceID int, props []model.DevicePropertyDetail) error {
	if len(props) == 0 || s.redisCmd == nil {
		return nil
	}

	values, err := s.redisCmd.HGetAll(ctx, fmt.Sprintf("device:%d:data", deviceID)).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return err
	}
	if len(values) == 0 {
		return nil
	}

	for i := range props {
		rawValue, ok := values[props[i].PropertyKey]
		if !ok {
			continue
		}
		parsed, parseErr := parseDevicePropValue(rawValue, props[i].DataType)
		if parseErr != nil {
			logger.Warn("parse device prop value failed",
				"device_id", deviceID,
				"property_key", props[i].PropertyKey,
				"raw_value", rawValue,
				"err", parseErr)
			continue
		}
		props[i].Value = parsed
	}

	return nil
}

func parseDevicePropValue(rawValue string, dataType model.ThingModelDataType) (any, error) {
	switch dataType {
	case model.ThingModelPropTypeBool:
		if rawValue == "1" {
			return true, nil
		}
		if rawValue == "0" {
			return false, nil
		}
		parsed, err := strconv.ParseBool(rawValue)
		if err != nil {
			return nil, err
		}
		return parsed, nil
	case model.ThingModelPropTypeInt:
		parsed, err := strconv.ParseInt(rawValue, 10, 64)
		if err != nil {
			return nil, err
		}
		return parsed, nil
	case model.ThingModelPropTypeFloat:
		parsed, err := strconv.ParseFloat(rawValue, 64)
		if err != nil {
			return nil, err
		}
		return parsed, nil
	default:
		return rawValue, nil
	}
}

func (s *DeviceSvc) getProductsMap(ctx context.Context, productIds []int) (map[int]model.Product, error) {
	models, err := s.productRepo.GetProductsByIds(ctx, productIds)
	if err != nil {
		return nil, err
	}

	var result = make(map[int]model.Product)
	for _, m := range models {
		result[m.Id] = m
	}
	return result, nil
}

// ---------------- 设备属性 ---------------- //

func (s *DeviceSvc) UpdateDeviceProp(ctx context.Context, req *dto.ReqDeviceProp) error {
	propDao, err := s.deviceRepo.GetDevicePropById(ctx, req.Id)
	if err != nil {
		return err
	}

	if err = s.deviceRepo.UpdateDevicePropRef(ctx, model.DevicePropertyRef{
		Id:         req.Id,
		Persistent: req.Persistent,
	}); err != nil {
		return err
	}

	if propDao.Persistent && !req.Persistent {
		// 移除设备缓存中的属性
		if err = notify.DevicePropChange(ctx, notify.OperationTypeDeleted, []*dataloader.DevicePropInfo{
			{
				DeviceId:   propDao.DeviceId,
				PropertyId: propDao.PropertyId,
			},
		}); err != nil {
			return err
		}
	}
	if !propDao.Persistent && req.Persistent {
		detail, err := s.deviceRepo.GetDevicePropDetailById(ctx, req.Id)
		if err != nil {
			return err
		}
		// 添加设备缓存中的属性
		if err = notify.DevicePropChange(ctx, notify.OperationTypeCreated, []*dataloader.DevicePropInfo{
			{
				DeviceId:    detail.DeviceId,
				DeviceKey:   detail.DeviceKey,
				PropertyId:  detail.PropertyId,
				PropertyKey: detail.PropertyKey,
			},
		}); err != nil {
			return err
		}
	}

	return err
}

func (s *DeviceSvc) DeleteDeviceProps(ctx context.Context, req *dto.ReqIds) error {
	return s.deviceRepo.DeleteDeviceProps(ctx, req.Ids)
}

func NewDeviceSvc(
	deviceRepo repo.IDeviceRepo,
	productRepo repo.IProductRepo,
	tmpRepo repo.IThingModelPropRepo,
	redisCmd redis.Cmdable,
) IDeviceSvc {
	return &DeviceSvc{
		deviceRepo:  deviceRepo,
		productRepo: productRepo,
		tmpRepo:     tmpRepo,
		redisCmd:    redisCmd,
	}
}
