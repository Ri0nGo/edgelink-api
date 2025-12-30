package repo

import (
	"context"
	"edgelink-api/internal/api/dto"
	"edgelink-api/internal/model"
	"edgelink-api/internal/pkg/paginate"

	"gorm.io/gorm"
)

type IThingModelPropRepo interface {
	CreateThingModelProp(ctx context.Context, tm *model.ThingModelProperty) error
	UpdateThingModelProp(ctx context.Context, tm *model.ThingModelProperty) error
	DeleteThingModelPropByModelId(ctx context.Context, modelId int) error
	GetThingModelPropsByModelId(ctx context.Context, modelId int) ([]model.ThingModelProperty, error)
	GetThingModelPropList(ctx context.Context, page dto.Page) (dto.Page, error)
}

type ThingModelPropRepo struct {
	db *gorm.DB
}

func (r *ThingModelPropRepo) GetThingModelPropsByModelId(ctx context.Context, modelId int) ([]model.ThingModelProperty, error) {
	var results []model.ThingModelProperty
	err := r.db.WithContext(ctx).
		Where("model_id = ?", modelId).
		Find(&results).
		Error
	return results, err
}

func (r *ThingModelPropRepo) CreateThingModelProp(ctx context.Context, tm *model.ThingModelProperty) error {
	err := r.db.
		WithContext(ctx).
		Create(&tm).
		Error
	return err
}

func (r *ThingModelPropRepo) UpdateThingModelProp(ctx context.Context, tm *model.ThingModelProperty) error {
	return r.db.
		WithContext(ctx).
		Where("id = ?", tm.Id).
		Updates(&tm).
		Error
}

func (r *ThingModelPropRepo) DeleteThingModelPropByModelId(ctx context.Context, modelId int) error {
	return r.db.
		WithContext(ctx).
		Where("model_id = ?", modelId).
		Delete(&model.ThingModel{}).
		Error
}

func (r *ThingModelPropRepo) GetThingModelPropList(ctx context.Context, page dto.Page) (dto.Page, error) {
	return paginate.PaginateList[model.ThingModel](ctx, r.db, page)
}

func NewThingModelPropRepo(db *gorm.DB) IThingModelPropRepo {
	return &ThingModelPropRepo{db: db}
}
