package svc

import (
	"context"
	"edgelink-api/internal/api/dto"
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
	// todo 新增属性后，需要查询该物模型是否有设备使用，如果有的话则需要同步至所有设备
	return s.tmpRepo.CreateThingModelProp(ctx, &model.ThingModelProperty{
		ModelId:    req.ModelId,
		Key:        req.Key,
		Name:       req.Name,
		Type:       req.Type,
		DataType:   req.DataType,
		Unit:       req.Unit,
		SourceType: req.SourceType,
		Expr:       req.Expr,
	})
}

func (s *ThingModelSvc) UpdateThingModelProp(ctx context.Context, req *dto.ReqThingModelProp) error {
	_, err := s.tmRepo.GetThingModelById(ctx, req.ModelId)
	if err != nil {
		return bizErr.NewBizError("模型不存在")
	}
	// todo 更新属性后，需要查询该物模型是否有设备使用，如果有的话则需要同步至所有设备
	return s.tmpRepo.UpdateThingModelProp(ctx, &model.ThingModelProperty{
		ModelId:    req.ModelId,
		Key:        req.Key,
		Name:       req.Name,
		Type:       req.Type,
		DataType:   req.DataType,
		Unit:       req.Unit,
		SourceType: req.SourceType,
		Expr:       req.Expr,
	})
}

func (s *ThingModelSvc) DeleteThingModelProp(ctx context.Context, id int) error {
	// todo 删除增属性后，需要查询该物模型是否有设备使用，如果有的话则需要同步至所有设备
	return s.tmpRepo.DeleteThingModelPropByModelId(ctx, id)
}

func (s *ThingModelSvc) GetThingModelPropList(ctx context.Context, modelId int, search string, page dto.Page) (dto.Page, error) {
	props, err := s.tmpRepo.GetThingModelPropsByModelId(ctx, modelId)
	if err != nil {
		return page, err
	}
	page.Data = props
	return page, nil
}

func NewThingModelSvc(tmRepo repo.IThingModelRepo, tmpRepo repo.IThingModelPropRepo, productRepo repo.IProductRepo) IThingModelSvc {
	return &ThingModelSvc{
		tmRepo:      tmRepo,
		tmpRepo:     tmpRepo,
		productRepo: productRepo,
	}
}
