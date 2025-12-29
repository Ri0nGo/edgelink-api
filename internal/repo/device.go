package repo

import (
	"context"
	"edgelink-api/internal/api/dto"
	"edgelink-api/internal/model"
	"edgelink-api/internal/pkg/paginate"

	"gorm.io/gorm"
)

type IDeviceRepo interface {
	CreateDevice(ctx context.Context, Device *model.Device) error
	UpdateDevice(ctx context.Context, Device *model.Device) error
	DeleteDevice(ctx context.Context, id int) error
	GetDeviceById(ctx context.Context, id int) (model.Device, error)
	GetDeviceList(ctx context.Context, page dto.Page) (dto.Page, error)
}

type DeviceRepo struct {
	db *gorm.DB
}

func (r *DeviceRepo) CreateDevice(ctx context.Context, Device *model.Device) error {
	return r.db.WithContext(ctx).Create(Device).Error
}

func (r *DeviceRepo) UpdateDevice(ctx context.Context, Device *model.Device) error {
	return r.db.WithContext(ctx).Where("id = ?", Device.Id).Updates(Device).Error
}

func (r *DeviceRepo) DeleteDevice(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(model.Device{}).Error
}

func (r *DeviceRepo) GetDeviceById(ctx context.Context, id int) (model.Device, error) {
	var result model.Device
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&result).Error
	return result, err
}

func (r *DeviceRepo) GetDeviceList(ctx context.Context, page dto.Page) (dto.Page, error) {
	return paginate.PaginateList[model.Device](ctx, r.db, page)
}

func NewDeviceRepo(db *gorm.DB) IDeviceRepo {
	return &DeviceRepo{db: db}
}
