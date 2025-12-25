package svc

import (
	"context"
	"edgelink-api/internal/api/dto"
	"edgelink-api/internal/model"
	"edgelink-api/internal/repo"
)

type IThingModelSvc interface {
	CreateThingModel(ctx context.Context, req *dto.ReqThingModel) error
	UpdateThingModel(ctx context.Context, req *dto.ReqThingModel) error
	DeleteThingModel(ctx context.Context, id int) error
	GetThingModelById(ctx context.Context, id int) (model.ThingModel, error)
	GetThingModelList(ctx context.Context, search string, page dto.Page) (dto.Page, error)
}

type ThingModelSvc struct {
	tmRepo repo.IThingModelRepo
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
	// todo 后面还需要检查是否有产品使用了该物模型，有则提示需要先处理物模型
	return s.tmRepo.DeleteThingModel(ctx, id)
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

func NewThingModelSvc(tmRepo repo.IThingModelRepo) IThingModelSvc {
	return &ThingModelSvc{
		tmRepo: tmRepo,
	}
}
