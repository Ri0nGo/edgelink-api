package dto

type ReqDevice struct {
	Id          int    `json:"id"`
	Name        string `json:"name" binding:"required"`       // 设备名称
	Key         string `json:"key"`                           // 设备唯一标识
	ProductId   int    `json:"product_id" binding:"required"` // 所属产品
	Description string `json:"description"`                   // 描述
}

type ReqDeviceProp struct {
	Id         int  `json:"id"`         // 属性id
	Persistent bool `json:"persistent"` // 持久化
	DeviceId   int  `json:"device_id"`
}
