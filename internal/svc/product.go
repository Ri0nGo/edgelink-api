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

type IProductSvc interface {
	CreateProduct(ctx context.Context, req *dto.ReqProduct) error
	UpdateProduct(ctx context.Context, req *dto.ReqProduct) error
	DeleteProduct(ctx context.Context, id int) error
	GetProductById(ctx context.Context, id int) (model.Product, error)
	GetProductList(ctx context.Context, search string, page dto.Page) (dto.Page, error)
}

type ProductSvc struct {
	tmRepo      repo.IThingModelRepo
	productRepo repo.IProductRepo
	deviceRepo  repo.IDeviceRepo
}

func (s *ProductSvc) CreateProduct(ctx context.Context, req *dto.ReqProduct) error {
	_, err := s.tmRepo.GetThingModelById(ctx, req.ModelId)
	if err != nil {
		return err
	}
	productDao := &model.Product{
		Identifier: req.Identifier,
		Name:       req.Name,
		ThingModelId: req.ModelId,
		Protocol:   req.Protocol,
	}

	return s.productRepo.CreateProduct(ctx, productDao)}

func (s *ProductSvc) UpdateProduct(ctx context.Context, req *dto.ReqProduct) error {
	_, err := s.tmRepo.GetThingModelById(ctx, req.ModelId)
	if err != nil {
		return err
	}
	productDao := &model.Product{
		Id:         req.Id,
		Identifier: req.Identifier,
		Name:       req.Name,
		ThingModelId: req.ModelId,
		Protocol:   req.Protocol,
	}
	return s.productRepo.UpdateProduct(ctx, productDao)
}

func (s *ProductSvc) DeleteProduct(ctx context.Context, id int) error {
	devices, err := s.deviceRepo.GetDevicesByProductId(ctx, id)
	if err != nil {
		return err
	}
	if len(devices) > 0 {
		return bizErr.NewBizError(fmt.Sprintf("该产品正在被 %s 设备使用，无法删除",
			utils.JoinByFunc(devices, ", ", func(device model.Device) string {
				return device.Name
			})))
	}
	return s.productRepo.DeleteProduct(ctx, id)
}

func (s *ProductSvc) GetProductById(ctx context.Context, id int) (model.Product, error) {
	productDao, err := s.productRepo.GetProductById(ctx, id)
	if err != nil {
		return model.Product{}, err
	}
	thingModel, err := s.tmRepo.GetThingModelById(ctx, productDao.ThingModelId)
	if err != nil {
		return model.Product{}, err
	}
	productDao.ModelName = thingModel.Name
	return productDao, nil
}

func (s *ProductSvc) GetProductList(ctx context.Context, search string, page dto.Page) (dto.Page, error) {
	dataPage, err := s.productRepo.GetProductList(ctx, search, page)
	if err != nil {
		return dto.Page{}, err
	}
	products, ok := dataPage.Data.([]model.Product)
	if !ok {
		return dto.Page{}, err
	}

	// 查询产品所使用的模型
	var modelIds []int
	for _, product := range products {
		modelIds = append(modelIds, product.ThingModelId)
	}

	modelsMap, err := s.getThingsModelsMap(ctx, modelIds)
	if err != nil {
		return dto.Page{}, err
	}
	for i, product := range products {
		if m, ok := modelsMap[product.ThingModelId]; ok {
			products[i].ModelName = m.Name
		}
	}
	dataPage.Data = products
	return dataPage, nil
}

func (s *ProductSvc) getThingsModelsMap(ctx context.Context, modelIds []int) (map[int]model.ThingModel, error) {
	models, err := s.tmRepo.GetThingModelsByIds(ctx, modelIds)
	if err != nil {
		return nil, err
	}

	var result = make(map[int]model.ThingModel)
	for _, m := range models {
		result[m.Id] = m
	}
	return result, nil
}

func NewProductSvc(productRepo repo.IProductRepo, tmRepo repo.IThingModelRepo, deviceRepo repo.IDeviceRepo) IProductSvc {
	return &ProductSvc{
		productRepo: productRepo,
		tmRepo:      tmRepo,
		deviceRepo:  deviceRepo,
	}
}
