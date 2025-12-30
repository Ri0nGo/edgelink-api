package repo

import (
	"context"
	"edgelink-api/internal/api/dto"
	"edgelink-api/internal/model"
	"edgelink-api/internal/pkg/paginate"

	"gorm.io/gorm"
)

type IDeviceRepo interface {
	CreateDevice(ctx context.Context, device *model.Device, props []model.DevicePropertyRef) error
	UpdateDevice(ctx context.Context, device *model.Device) error
	DeleteDevice(ctx context.Context, id int) error
	GetDeviceById(ctx context.Context, id int) (model.Device, error)
	GetDevicesByProductId(ctx context.Context, productId int) ([]model.Device, error)
	GetDeviceList(ctx context.Context, page dto.Page) (dto.Page, error)

	// device property
	UpdateDevicePropRef(ctx context.Context, props model.DevicePropertyRef) error
	DeleteDeviceProps(ctx context.Context, id []int) error
	DeleteDevicePropByDeviceId(ctx context.Context, deviceId int) error
}

type DeviceRepo struct {
	db *gorm.DB
}

func (r *DeviceRepo) GetDevicesByProductId(ctx context.Context, productId int) ([]model.Device, error) {
	var results []model.Device
	err := r.db.WithContext(ctx).
		Where("product_id = ?", productId).
		Find(&results).
		Error
	return results, err
}

func (r *DeviceRepo) CreateDevice(ctx context.Context, device *model.Device, props []model.DevicePropertyRef) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(device).Error; err != nil {
			return err
		}

		if len(props) > 0 {
			for i := range props {
				props[i].DeviceId = device.Id
			}
			if err := tx.Model(&model.DevicePropertyRef{}).Create(&props).Error; err != nil {
				return err
			}
		}

		return nil
	})
	return err
}

func (r *DeviceRepo) UpdateDevice(ctx context.Context, device *model.Device) error {
	return r.db.WithContext(ctx).Where("id = ?", device.Id).Updates(device).Error
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

func (r *DeviceRepo) UpdateDevicePropRef(ctx context.Context, prop model.DevicePropertyRef) error {
	err := r.db.WithContext(ctx).
		Where("id = ?", prop.Id).
		Updates(prop).
		Error
	return err
}

func (r *DeviceRepo) DeleteDeviceProps(ctx context.Context, ids []int) error {
	err := r.db.WithContext(ctx).
		Where("id IN ?", ids).
		Delete(&model.DevicePropertyRef{}).
		Error
	return err
}

func (r *DeviceRepo) DeleteDevicePropByDeviceId(ctx context.Context, deviceId int) error {
	err := r.db.WithContext(ctx).
		Where("device_id = ?", deviceId).
		Delete(&model.DevicePropertyRef{}).
		Error
	return err
}

func NewDeviceRepo(db *gorm.DB) IDeviceRepo {
	return &DeviceRepo{db: db}
}
