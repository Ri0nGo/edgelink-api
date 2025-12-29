package svc

import (
	"context"
	"edgelink-api/internal/api/dto"
	"edgelink-api/internal/model"
	bizErr "edgelink-api/internal/pkg/bizerr"
	"edgelink-api/internal/repo"
	"errors"
	"fmt"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type IDeviceSvc interface {
	CreateDevice(ctx context.Context, req *dto.ReqDevice) error
	UpdateDevice(ctx context.Context, req *dto.ReqDevice) error
	DeleteDevice(ctx context.Context, id int) error
	GetDeviceById(ctx context.Context, id int) (model.Device, error)
	GetDeviceList(ctx context.Context, page dto.Page) (dto.Page, error)
}

type DeviceSvc struct {
	productRepo repo.IProductRepo
	deviceRepo  repo.IDeviceRepo
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

	return s.deviceRepo.CreateDevice(ctx, DeviceDao)
}

func (s *DeviceSvc) generateAddress(productKey, deviceKey string) model.DeviceAddress {
	return model.DeviceAddress{
		Uplink: []model.DeviceAddressDetail{
			{
				Address: fmt.Sprintf("/devices/%s/%s/uplink/data}", productKey, deviceKey),
				Desc:    "上传设备数据至MQTT",
			}, {
				Address: fmt.Sprintf("/devices/%s/%s/uplink/status}", productKey, deviceKey),
				Desc:    "上传设备状态至MQTT",
			}, {
				Address: fmt.Sprintf("/devices/%s/%s/uplink/keepalive}", productKey, deviceKey),
				Desc:    "同步设备心跳",
			},
		}, Downlink: []model.DeviceAddressDetail{
			{
				Address: fmt.Sprintf("/devices/%s/%s/downlink/event}", productKey, deviceKey),
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
	return s.deviceRepo.DeleteDevice(ctx, id)
}

func (s *DeviceSvc) GetDeviceById(ctx context.Context, id int) (model.Device, error) {
	DeviceDao, err := s.deviceRepo.GetDeviceById(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.Device{}, bizErr.NewBizError("设备不存在")
		}
		return model.Device{}, err
	}
	productModel, err := s.productRepo.GetProductById(ctx, DeviceDao.ProductId)
	if err != nil {
		return model.Device{}, err
	}
	DeviceDao.ProductName = productModel.Name
	return DeviceDao, nil
}

func (s *DeviceSvc) GetDeviceList(ctx context.Context, page dto.Page) (dto.Page, error) {
	dataPage, err := s.deviceRepo.GetDeviceList(ctx, page)
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
	dataPage.Data = Devices
	return dataPage, nil
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

func NewDeviceSvc(deviceRepo repo.IDeviceRepo, productRepo repo.IProductRepo) IDeviceSvc {
	return &DeviceSvc{
		deviceRepo:  deviceRepo,
		productRepo: productRepo,
	}
}
