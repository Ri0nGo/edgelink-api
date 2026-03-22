package model

import "time"

type ThingModelDataType int8

const (
	ThingModelPropTypeBool ThingModelDataType = iota + 1
	ThingModelPropTypeInt
	ThingModelPropTypeFloat
)

type ThingModelSourceType int8 // 属性

const (
	ThingModelSourceTypeRaw ThingModelSourceType = iota + 1
	ThingModelSourceTypeFormula
)

type ThingModel struct {
	Id          int                  `json:"id" gorm:"primaryKey;autoIncrement;comment:支付自增ID"`
	Identifier  string               `json:"identifier" gorm:"type:varchar(64);not null;uniqueIndex:uk_model_identifier;comment:模型标识符"`
	Name        string               `json:"name" gorm:"type:varchar(128);not null;comment:模型名称"`
	Description string               `json:"description" gorm:"type:varchar(255);not null;comment:描述"`
	Icon        string               `json:"icon"`
	CreatedTime time.Time            `json:"created_time" gorm:"autoCreateTime"`
	UpdatedTime time.Time            `json:"updated_time" gorm:"autoUpdateTime"`
	Props       []ThingModelProperty `json:"props" gorm:"-"`
}

func (m ThingModel) TableName() string {
	return "thing_model"
}

type ThingModelType int8

const (
	ThingModelPropType ThingModelType = iota + 1
	ThingModelFuncType
	ThingModelEventType
)

type ThingModelProperty struct {
	Id          int                  `json:"id" gorm:"primaryKey;autoIncrement;comment:模型属性id"`
	ModelId     int                  `json:"model_id"`
	Key         string               `json:"key" gorm:"type:varchar(64);not null;uniqueIndex:uk_model_id_key;comment:数据key"`
	Name        string               `json:"name"`
	Type        ThingModelType       `json:"type"` // 类型 `jsunit"
	DataType    ThingModelDataType   `json:"data_type"`
	Unit        string               `json:"unit"`
	SourceType  ThingModelSourceType `json:"source_type" gorm:"default:1;comment:数据来源 1=原始数据 2=公式计算"`
	Expr        string               `json:"expr" gorm:"comment:表达式"`
	CreatedTime time.Time            `json:"created_time" gorm:"autoCreateTime"`
	UpdatedTime time.Time            `json:"updated_time" gorm:"autoUpdateTime"`
}

func (m ThingModelProperty) TableName() string {
	return "thing_model_property"
}
