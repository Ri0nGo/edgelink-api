package bootstrap

import (
	"context"
	"edgelink-api/internal/dataloader"
	"edgelink-api/internal/dataloader/notify"
	"edgelink-api/internal/pkg/logger"

	"gorm.io/gorm"
)

func PublishDeviceConfigToRedis(ctx context.Context, db *gorm.DB, pub notify.NotifierPub) error {
	deviceInfos, err := getDeviceConfigs(ctx, db)
	if err != nil {
		return err
	}
	for _, info := range deviceInfos {
		err = pub.DeviceConfigChange(ctx, notify.DeviceNotifyType, notify.OperationTypeCreated, &info)
		if err != nil {
			return err
		}
	}
	logger.Info("publish device configs to redis", "length", len(deviceInfos))
	return nil
}

func getDeviceConfigs(ctx context.Context, db *gorm.DB) ([]dataloader.DeviceInfo, error) {
	var deviceInfos []dataloader.DeviceInfo

	if err := db.WithContext(ctx).
		Raw(`
            SELECT 
                t1.id AS device_id, 
                t1.key AS device_key, 
                t2.identifier AS product_identifier 
            FROM device t1 
            JOIN product t2 ON t1.product_id = t2.id
        `).
		Scan(&deviceInfos).Error; err != nil {
		return nil, err
	}

	return deviceInfos, nil
}

func PublishDevicePropsToRedis(ctx context.Context, db *gorm.DB, pub notify.NotifierPub) error {
	props, err := getDevicePropInfos(ctx, db)
	if err != nil {
		return err
	}

	devicePropM := make(map[int][]*dataloader.DevicePropInfo)
	for _, prop := range props {
		devicePropM[prop.DeviceId] = append(devicePropM[prop.DeviceId], prop)
	}

	for _, deviceProps := range devicePropM {
		err = pub.DevicePropChange(ctx, notify.DevicePropertyNotifyType, notify.OperationTypeCreated, deviceProps)
		if err != nil {
			return err
		}
		//logger.Info("notify device props success", "device_id", deviceId, "length", len(deviceProps))
	}
	logger.Info("publish device props to redis")
	return nil
}

func getDevicePropInfos(ctx context.Context, db *gorm.DB) ([]*dataloader.DevicePropInfo, error) {
	var deviceInfos []*dataloader.DevicePropInfo

	if err := db.WithContext(ctx).
		Raw(`
SELECT t2.id device_id, t2.key as device_key, t3.id as property_id, t3.key as property_key from device_property_ref t1
INNER JOIN device t2
INNER JOIN thing_model_property t3
ON t1.device_id = t2.id AND t1.property_id = t3.id;
`).
		Scan(&deviceInfos).Error; err != nil {
		return nil, err
	}

	return deviceInfos, nil
}
