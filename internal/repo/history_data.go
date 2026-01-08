package repo

import "gorm.io/gorm"

type IHistoryDataRepo interface {
	DeleteHistoryDataByDeviceId(deviceId int) error
	DeleteHistoryDataByPropertyId(propertyId int) error
}
type HistoryDataRepo struct {
	db *gorm.DB
}

func (r *HistoryDataRepo) DeleteHistoryDataByDeviceId(deviceId int) error {
	//TODO implement me
	panic("implement me")
}

func (r *HistoryDataRepo) DeleteHistoryDataByPropertyId(propertyId int) error {
	//TODO implement me
	panic("implement me")
}
