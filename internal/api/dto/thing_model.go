package dto

import (
	"edgelink-api/internal/model"
)

type FuncType struct {
	Id          int                        `json:"id"`
	Name        string                     `json:"name" binding:"required"`
	ModelId     int                        `json:"model_id" binding:"required"`
	Key         string                     `json:"Key" binding:"required"` // 属性标识
	Description string                     `json:"description"`
	SourceType  model.ThingModelSourceType `json:"source_type"`
	Type        model.ThingModelType       `json:"type"`
	DataType    model.ThingModelDataType   `json:"data_type"`
	Expr        string                     `json:"expr"` // 公式表达式，这一版先不考虑
	Unit        string                     `json:"unit"`
}

type ReqThingModel struct {
	Id          int        `json:"id"`
	Name        string     `json:"name" binding:"required"`
	Identifier  string     `json:"identifier" binding:"required"` // 标识符
	Icon        string     `json:"icon"`
	Description string     `json:"description"`
	FuncTypes   []FuncType `json:"func_types"`
}

type RespThingModelList struct {
	Page
	Search string `json:"search" form:"search"`
}
