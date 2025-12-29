package repo

import (
	"context"
	"edgelink-api/internal/api/dto"
	"edgelink-api/internal/model"
	"edgelink-api/internal/pkg/paginate"

	"gorm.io/gorm"
)

type IProductRepo interface {
	CreateProduct(ctx context.Context, product *model.Product) error
	UpdateProduct(ctx context.Context, product *model.Product) error
	DeleteProduct(ctx context.Context, id int) error
	GetProductById(ctx context.Context, id int) (model.Product, error)
	GetProductsByIds(ctx context.Context, ids []int) ([]model.Product, error)
	GetProductList(ctx context.Context, page dto.Page) (dto.Page, error)
}

type ProductRepo struct {
	db *gorm.DB
}

func (r *ProductRepo) GetProductsByIds(ctx context.Context, ids []int) ([]model.Product, error) {
	var results []model.Product
	err := r.db.WithContext(ctx).
		Find(&results).
		Error
	return results, err
}

func (r *ProductRepo) CreateProduct(ctx context.Context, product *model.Product) error {
	return r.db.WithContext(ctx).Create(product).Error
}

func (r *ProductRepo) UpdateProduct(ctx context.Context, product *model.Product) error {
	return r.db.WithContext(ctx).Where("id = ?", product.Id).Updates(product).Error
}

func (r *ProductRepo) DeleteProduct(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(model.Product{}).Error
}

func (r *ProductRepo) GetProductById(ctx context.Context, id int) (model.Product, error) {
	var result model.Product
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&result).Error
	//err := r.db.WithContext(ctx).
	//	Table("product as t1").
	//	Select("t1.*, t2.name as model_name").
	//	Joins("inner join thing_model t2 on t1.model_id = t2.id").
	//	Where("t1.id = ?", id).
	//	Scan(&result).
	//	Error
	return result, err
}

func (r *ProductRepo) GetProductList(ctx context.Context, page dto.Page) (dto.Page, error) {
	return paginate.PaginateList[model.Product](ctx, r.db, page)
}

func NewProductRepo(db *gorm.DB) IProductRepo {
	return &ProductRepo{db: db}
}
