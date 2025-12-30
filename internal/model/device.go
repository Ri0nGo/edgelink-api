package model

import (
	"time"

	"gorm.io/datatypes"
)

type DeviceAddressDetail struct {
	Address string `json:"address"`
	Desc    string `json:"desc"`
}

type DeviceAddress struct {
	Uplink   []DeviceAddressDetail `json:"uplink"`   // 上行
	Downlink []DeviceAddressDetail `json:"downlink"` // 下行
}

type Device struct {
	Id          int                               `json:"id" gorm:"primaryKey;autoIncrement"`
	Key         string                            `json:"device_key"`
	Name        string                            `json:"device_name"`
	ProductId   int                               `json:"product_id"`
	ProductName string                            `json:"product_name" gorm:"-"`
	Address     datatypes.JSONType[DeviceAddress] `json:"address"`
	Description string                            `json:"description"`
	CreatedTime time.Time                         `json:"created_time" gorm:"autoCreateTime"`
	UpdatedTime time.Time                         `json:"updated_time" gorm:"autoUpdateTime"`
}

func (d Device) TableName() string {
	return "device"
}

type StoreMode string

const (
	StoreModeFull   StoreMode = "full"
	StoreModeChange StoreMode = "change"
	StoreModeMinute StoreMode = "minute"
	StoreModeHour   StoreMode = "hour"
	StoreModeDay    StoreMode = "day"
	StoreModeWeek   StoreMode = "week"
	StoreModeMonth  StoreMode = "month"
)

type DevicePropertyRef struct {
	Id         int `json:"id" gorm:"primaryKey;autoIncrement"`
	DeviceId   int `json:"device_id"`
	PropertyId int `json:"property_id"`
	// 数据持久化
	Persistent bool `json:"persistent"`
	// 数据存储方式，full全量存储(每接收到一个值就存下来), change值变化才存储, minute每分钟一条
	StoreMode   StoreMode `json:"store_mode" gorm:"default:minute"` // todo 待思考是否有必要使用该字段
	CreatedTime time.Time `json:"created_time" gorm:"autoCreateTime"`
	UpdatedTime time.Time `json:"updated_time" gorm:"autoUpdateTime"`
}

func (d DevicePropertyRef) TableName() string {
	return "device_property_ref"
}
