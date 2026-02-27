package model

import "time"

type HistoryData struct {
	DeviceId   int       `json:"device_id"`
	PropertyId int       `json:"property_id"`
	Ts         time.Time `json:"ts"`
	Value      float64   `json:"value"`
}

func (d *HistoryData) TableName() string {
	return "history_data"
}
