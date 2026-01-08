package svc

import (
	"context"
	"edgelink-api/internal/api/dto"
	"edgelink-api/internal/dataloader"
	"edgelink-api/internal/dataloader/notify"
	"edgelink-api/internal/model"
	bizErr "edgelink-api/internal/pkg/bizerr"
	"edgelink-api/internal/repo"
	"edgelink-api/internal/utils"
	"fmt"
)

type IThingModelSvc interface {
	CreateThingModel(ctx context.Context, req *dto.ReqThingModel) error
	UpdateThingModel(ctx context.Context, req *dto.ReqThingModel) error
	DeleteThingModel(ctx context.Context, id int) error
	GetThingModelById(ctx context.Context, id int) (model.ThingModel, error)
	GetThingModelList(ctx context.Context, search string, page dto.Page) (dto.Page, error) //

	// things model prop
	CreateThingModelProp(ctx context.Context, req *dto.ReqThingModelProp) error
	UpdateThingModelProp(ctx context.Context, req *dto.ReqThingModelProp) error
	DeleteThingModelProp(ctx context.Context, id int) error
	GetThingModelPropList(ctx context.Context, modelId int, search string, page dto.Page) (dto.Page, error)
}

type ThingModelSvc struct {
	tmRepo      repo.IThingModelRepo
	tmpRepo     repo.IThingModelPropRepo
	productRepo repo.IProductRepo
	deviceRepo  repo.IDeviceRepo
}

func (s *ThingModelSvc) CreateThingModel(ctx context.Context, req *dto.ReqThingModel) error {
	tmDao := &model.ThingModel{
		Identifier:  req.Identifier,
		Name:        req.Name,
		Description: req.Description,
		Icon:        req.Icon,
	}

	// 更新模型功能定义
	tmps := make([]model.ThingModelProperty, 0, len(req.FuncTypes))
	for _, funcType := range req.FuncTypes {
		tmps = append(tmps, model.ThingModelProperty{
			Key:        funcType.Key,
			Name:       funcType.Name,
			Type:       funcType.Type,
			DataType:   funcType.DataType,
			Unit:       funcType.Unit,
			SourceType: funcType.SourceType,
			Expr:       funcType.Expr,
		})
	}

	err := s.tmRepo.CreateThingModel(ctx, tmDao, tmps)
	return err
}

func (s *ThingModelSvc) UpdateThingModel(ctx context.Context, req *dto.ReqThingModel) error {
	tmDao := &model.ThingModel{
		Id:          req.Id,
		Identifier:  req.Identifier,
		Name:        req.Name,
		Description: req.Description,
		Icon:        req.Icon,
	}

	// 更新模型功能定义
	tmps := make([]model.ThingModelProperty, 0, len(req.FuncTypes))
	for _, funcType := range req.FuncTypes {
		tmps = append(tmps, model.ThingModelProperty{
			Id:         funcType.Id,
			ModelId:    funcType.ModelId,
			Key:        funcType.Key,
			Name:       funcType.Name,
			Type:       funcType.Type,
			DataType:   funcType.DataType,
			Unit:       funcType.Unit,
			SourceType: funcType.SourceType,
			Expr:       funcType.Expr,
		})
	}

	err := s.tmRepo.UpdateThingModel(ctx, tmDao, tmps)
	return err
}

func (s *ThingModelSvc) DeleteThingModel(ctx context.Context, id int) error {
	products, err := s.productRepo.GetProductsByThingModelId(ctx, id)
	if err != nil {
		return err
	}
	if len(products) > 0 {
		return bizErr.NewBizError(fmt.Sprintf("该模型正在被 %s 产品使用，无法删除",
			utils.JoinByFunc(products, ", ", func(product model.Product) string {
				return product.Name
			})))
	}
	if err = s.tmRepo.DeleteThingModel(ctx, id); err != nil {
		return err
	}
	return s.tmpRepo.DeleteThingModelPropByModelId(ctx, id)
}

func (s *ThingModelSvc) GetThingModelById(ctx context.Context, id int) (model.ThingModel, error) {
	tmDao, err := s.tmRepo.GetThingModelById(ctx, id)
	if err != nil {
		return model.ThingModel{}, err
	}
	return tmDao, nil
}

func (s *ThingModelSvc) GetThingModelList(ctx context.Context, search string, page dto.Page) (dto.Page, error) {
	return s.tmRepo.GetThingModelList(ctx, search, page)
}

// ---------------- 物模型属性 ---------------- //

func (s *ThingModelSvc) CreateThingModelProp(ctx context.Context, req *dto.ReqThingModelProp) error {
	_, err := s.tmRepo.GetThingModelById(ctx, req.ModelId)
	if err != nil {
		return bizErr.NewBizError("模型不存在")
	}

	propDao := &model.ThingModelProperty{
		ModelId:    req.ModelId,
		Key:        req.Key,
		Name:       req.Name,
		Type:       req.Type,
		DataType:   req.DataType,
		Unit:       req.Unit,
		SourceType: req.SourceType,
		Expr:       req.Expr,
	}
	if err = s.tmpRepo.CreateThingModelProp(ctx, propDao); err != nil {
		return err
	}

	// 查询使用了该物模型的所有设备
	devices, err := s.deviceRepo.GetDevicesByThingModelID(ctx, req.ModelId)
	if err != nil {
		return err
	}

	var (
		deviceProps    = make([]*dataloader.DevicePropInfo, len(devices))
		devicePropRefs = make([]model.DevicePropertyRef, len(devices))
	)
	for idx, device := range devices {
		deviceProps[idx] = &dataloader.DevicePropInfo{
			DeviceId:    device.Id,
			DeviceKey:   device.Key,
			PropertyId:  propDao.Id,
			PropertyKey: propDao.Key,
		}
		devicePropRefs[idx] = model.DevicePropertyRef{
			Id:         0,
			DeviceId:   device.Id,
			PropertyId: propDao.Id,
			Persistent: true,
			StoreMode:  model.StoreModeMinute,
		}
	}
	// 同步属性到使用了该模型的所有设备
	if err = s.deviceRepo.CreateDevicePropRef(ctx, devicePropRefs); err != nil {
		return err
	}

	// 通知存储器更新缓存
	if err = notify.DevicePropChange(ctx, notify.OperationTypeCreated, deviceProps); err != nil {
		return err
	}
	return nil
}

func (s *ThingModelSvc) UpdateThingModelProp(ctx context.Context, req *dto.ReqThingModelProp) error {
	_, err := s.tmRepo.GetThingModelById(ctx, req.ModelId)
	if err != nil {
		return bizErr.NewBizError("模型不存在")
	}

	propDao := &model.ThingModelProperty{
		Id:         req.Id,
		ModelId:    req.ModelId,
		Key:        req.Key,
		Name:       req.Name,
		Type:       req.Type,
		DataType:   req.DataType,
		Unit:       req.Unit,
		SourceType: req.SourceType,
		Expr:       req.Expr,
	}
	if err = s.tmpRepo.UpdateThingModelProp(ctx, propDao); err != nil {
		return err
	}

	// 查询使用了该物模型的所有设备
	devices, err := s.deviceRepo.GetDevicesByThingModelID(ctx, req.ModelId)
	if err != nil {
		return err
	}

	var deviceProps = make([]*dataloader.DevicePropInfo, len(devices))
	for idx, device := range devices {
		deviceProps[idx] = &dataloader.DevicePropInfo{
			DeviceId:    device.Id,
			DeviceKey:   device.Key,
			PropertyId:  propDao.Id,
			PropertyKey: propDao.Key,
		}
	}

	// 通知存储器更新缓存
	if err = notify.DevicePropChange(ctx, notify.OperationTypeUpdated, deviceProps); err != nil {
		return err
	}
	return nil
}

func (s *ThingModelSvc) DeleteThingModelProp(ctx context.Context, id int) error {
	if err := s.tmpRepo.DeleteThingModelProp(ctx, id); err != nil {
		return err
	}
	props, err := s.deviceRepo.GetDevicePropsByPropId(ctx, id)
	if err != nil {
		return err
	}
	var deviceProps = make([]*dataloader.DevicePropInfo, len(props))
	for i, prop := range props {
		deviceProps[i] = &dataloader.DevicePropInfo{
			DeviceId:   prop.DeviceId,
			PropertyId: prop.PropertyId,
		}
	}
	if err = notify.DevicePropChange(ctx, notify.OperationTypeDeleted, deviceProps); err != nil {
		return err
	}
	return nil
}

func (s *ThingModelSvc) GetThingModelPropList(ctx context.Context, modelId int, search string, page dto.Page) (dto.Page, error) {
	props, err := s.tmpRepo.GetThingModelPropsByModelId(ctx, modelId)
	if err != nil {
		return page, err
	}
	page.Data = props
	return page, nil
}

func NewThingModelSvc(tmRepo repo.IThingModelRepo, tmpRepo repo.IThingModelPropRepo,
	productRepo repo.IProductRepo, deviceRepo repo.IDeviceRepo) IThingModelSvc {
	return &ThingModelSvc{
		tmRepo:      tmRepo,
		tmpRepo:     tmpRepo,
		productRepo: productRepo,
		deviceRepo:  deviceRepo,
	}
}
