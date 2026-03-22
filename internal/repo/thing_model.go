package repo

import (
	"context"
	"edgelink-api/internal/api/dto"
	"edgelink-api/internal/model"
	"edgelink-api/internal/pkg/paginate"

	"gorm.io/gorm"
)

type IThingModelRepo interface {
	CreateThingModel(ctx context.Context, tm *model.ThingModel, tmps []model.ThingModelProperty) error
	UpdateThingModel(ctx context.Context, tm *model.ThingModel, tmps []model.ThingModelProperty) error
	DeleteThingModel(ctx context.Context, id int) error
	GetThingModelById(ctx context.Context, id int) (model.ThingModel, error)
	GetThingModelsByIds(ctx context.Context, ids []int) ([]model.ThingModel, error)
	GetThingModelList(ctx context.Context, search string, page dto.Page) (dto.Page, error)
}

type ThingModelRepo struct {
	db *gorm.DB
}

func (r *ThingModelRepo) CreateThingModel(ctx context.Context, tm *model.ThingModel, tmps []model.ThingModelProperty) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(tm).Error; err != nil {
			return err
		}
		if len(tmps) == 0 {
			return nil
		}

		for i := range tmps {
			tmps[i].ModelId = tm.Id
		}
		if err := tx.Create(&tmps).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *ThingModelRepo) UpdateThingModel(ctx context.Context, tm *model.ThingModel, tmps []model.ThingModelProperty) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ?", tm.Id).Updates(tm).Error; err != nil {
			return err
		}
		for _, tmp := range tmps {
			if err := tx.Where("id = ?", tmp.Id).Updates(&tmp).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *ThingModelRepo) DeleteThingModel(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ?", id).Delete(&model.ThingModel{}).Error; err != nil {
			return err
		}
		if err := tx.Where("model_id = ?", id).Delete(&model.ThingModelProperty{}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *ThingModelRepo) GetThingModelById(ctx context.Context, id int) (model.ThingModel, error) {
	var tm model.ThingModel
	err := r.db.
		WithContext(ctx).
		Where("id = ?", id).
		First(&tm).
		Error
	return tm, err
}

func (r *ThingModelRepo) GetThingModelsByIds(ctx context.Context, ids []int) ([]model.ThingModel, error) {
	var tms []model.ThingModel
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&tms).Error
	return tms, err
}

func (r *ThingModelRepo) GetThingModelList(ctx context.Context, search string, page dto.Page) (dto.Page, error) {
	return paginate.PaginateList[model.ThingModel](ctx, r.db, page, func(db *gorm.DB) *gorm.DB {
		if search != "" {
			db = db.Where("name LIKE ? OR identifier LIKE ?", "%"+search+"%", "%"+search+"%")
		}
		return db
	})
}

func NewThingModelRepo(db *gorm.DB) IThingModelRepo {
	return &ThingModelRepo{db: db}
}
