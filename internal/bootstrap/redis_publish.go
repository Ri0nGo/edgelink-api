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
