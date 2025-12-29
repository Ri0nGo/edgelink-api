package repo

import (
	"context"
	"edgelink-api/internal/model"
	"gorm.io/gorm"
)

type IDevicePropRefRepo interface {
	CreateDevicePropRef(ctx context.Context, DevicePropRef *model.DevicePropertyRef) error
	UpdateDevicePropRef(ctx context.Context, DevicePropRef *model.DevicePropertyRef) error
	DeleteDevicePropRef(ctx context.Context, id int) error
	GetDevicePropRefsByDeviceId(ctx context.Context, deviceId int) ([]model.DevicePropertyRef, error)
}

type DevicePropRefRepo struct {
	db *gorm.DB
}

func (r *DevicePropRefRepo) CreateDevicePropRef(ctx context.Context, DevicePropRef *model.DevicePropertyRef) error {
	return r.db.WithContext(ctx).Create(DevicePropRef).Error
}

func (r *DevicePropRefRepo) UpdateDevicePropRef(ctx context.Context, DevicePropRef *model.DevicePropertyRef) error {
	return r.db.WithContext(ctx).Where("id = ?", DevicePropRef.Id).Updates(DevicePropRef).Error
}

func (r *DevicePropRefRepo) DeleteDevicePropRef(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(model.DevicePropertyRef{}).Error
}

func (r *DevicePropRefRepo) GetDevicePropRefsByDeviceId(ctx context.Context, deviceId int) ([]model.DevicePropertyRef, error) {
	var result []model.DevicePropertyRef
	err := r.db.WithContext(ctx).
		Where("device_id = ?", deviceId).
		First(&result).Error
	return result, err
}

//func (r *DevicePropRefRepo) GetDevicePropRefList(ctx context.Context, page dto.Page) (dto.Page, error) {
//	return paginate.PaginateList[model.DevicePropertyRef](ctx, r.db, page)
//}

func NewDevicePropRefRepo(db *gorm.DB) IDevicePropRefRepo {
	return &DevicePropRefRepo{db: db}
}
