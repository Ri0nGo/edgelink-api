package model

import "time"

type ProductProtocolType string

const (
	ProductProtocolTypeMQTT ProductProtocolType = "mqtt"
)

type Product struct {
	Id          int                 `json:"id" gorm:"primaryKey;autoIncrement"`
	Identifier  string              `json:"identifier" gorm:"type:varchar(64);not null;uniqueIndex:uk_product_identifier"` // 产品标识符
	Name        string              `json:"name"`
	ThingModelId int                 `json:"model_id"` // 对应表列 thing_model_id；json tag 保持 model_id 兼容前端
	ModelName   string              `json:"model_name" gorm:"-"`
	Protocol    ProductProtocolType `json:"protocol"` // 产品使用的协议
	CreatedTime time.Time           `json:"created_time" gorm:"autoCreateTime"`
	UpdatedTime time.Time           `json:"updated_time" gorm:"autoUpdateTime"`
}

func (m Product) TableName() string {
	return "product"
}
