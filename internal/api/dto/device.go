package dto

import "edgelink-api/internal/model"

type ReqDevice struct {
	Id          int    `json:"id"`
	Name        string `json:"name" binding:"required"`       // 设备名称
	Key         string `json:"key"`                           // 设备唯一标识
	ProductId   int    `json:"product_id" binding:"required"` // 所属产品
	Description string `json:"description"`                   // 描述
}

type ReqDeviceProp struct {
	Id         int  `json:"id"`         // 设备属性关系表的 id
	Persistent bool `json:"persistent"` // 持久化
	DeviceId   int  `json:"device_id"`
}

type RespDevice struct {
	model.Device
	Props []model.DevicePropertyDetail `json:"props"` // 属性， 功能和事件还未实现
}
