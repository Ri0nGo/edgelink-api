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
	GetDevicesByThingModelID(ctx context.Context, modelId int) ([]model.Device, error)

	// device property
	UpdateDevicePropRef(ctx context.Context, props model.DevicePropertyRef) error
	DeleteDeviceProps(ctx context.Context, id []int) error
	DeleteDevicePropByDeviceId(ctx context.Context, deviceId int) error
	GetDevicePropByDeviceId(ctx context.Context, deviceId int) ([]model.DevicePropertyDetail, error)
	GetDevicePropsByPropId(ctx context.Context, propId int) ([]model.DevicePropertyRef, error)
	GetDevicePropById(ctx context.Context, id int) (model.DevicePropertyRef, error)
	GetDevicePropDetailById(ctx context.Context, id int) (model.DevicePropertyDetail, error)
	CreateDevicePropRef(ctx context.Context, props []model.DevicePropertyRef) error
}

type DeviceRepo struct {
	db *gorm.DB
}

// GetDevicePropDetailById 获取设备及物模型属性详情
func (r *DeviceRepo) GetDevicePropDetailById(ctx context.Context, id int) (model.DevicePropertyDetail, error) {
	var prop model.DevicePropertyDetail
	err := r.db.WithContext(ctx).
		Raw(`
SELECT
    t1.id, t1.persistent, t1.store_mode,
    t2.id as property_id, t2.key as property_key, t2.name as property_name, t2.data_type, t2.unit, t2.source_type, t2.expr, t2.type,
    t3.id as device_id, t3.key as device_key
FROM device_property_ref t1
INNER JOIN thing_model_property t2
INNER JOIN device t3
ON t1.property_id = t2.id AND t1.device_id = t3.id
WHERE t1.id
`, id).
		First(&prop).
		Error
	return prop, err
}

func (r *DeviceRepo) GetDevicePropById(ctx context.Context, id int) (model.DevicePropertyRef, error) {
	var result model.DevicePropertyRef
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&result).
		Error
	return result, err
}

// GetDevicePropsByPropId 通过属性id查询
func (r *DeviceRepo) GetDevicePropsByPropId(ctx context.Context, propId int) ([]model.DevicePropertyRef, error) {
	var results []model.DevicePropertyRef
	err := r.db.WithContext(ctx).
		Where("property_id = ?", propId).
		Find(&results).
		Error
	return results, err
}

func (r *DeviceRepo) CreateDevicePropRef(ctx context.Context, props []model.DevicePropertyRef) error {
	return r.db.WithContext(ctx).
		Create(props).
		Error
}

// GetDevicesByThingModelID 通过物模型id查询设备列表
func (r *DeviceRepo) GetDevicesByThingModelID(ctx context.Context, modelId int) ([]model.Device, error) {
	var results []model.Device
	err := r.db.WithContext(ctx).
		Raw(`
SELECT t1.* from device t1
INNER JOIN product t2
INNER JOIN thing_model t3
ON t1.product_id = t2.id AND t2.thing_model_id = t3.id
WHERE t3.id = ?;
`, modelId).
		Find(&results).
		Error
	return results, err
}

// GetDevicePropByDeviceId 查询设备使用了哪些模型属性
func (r *DeviceRepo) GetDevicePropByDeviceId(ctx context.Context, deviceId int) ([]model.DevicePropertyDetail, error) {
	var props []model.DevicePropertyDetail
	err := r.db.WithContext(ctx).
		Raw(`
SELECT t1.id, t1.persistent, t1.store_mode, t2.id as property_id, t2.key as property_key, t2.name as property_name, t2.data_type, t2.unit, t2.source_type, t2.expr, t2.type 
FROM device_property_ref t1
INNER JOIN thing_model_property t2
ON t1.property_id = t2.id
WHERE t1.device_id = ?
`, deviceId).
		Find(&props).
		Error
	return props, err
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
