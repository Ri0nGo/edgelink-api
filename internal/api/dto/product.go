package dto

import (
	"edgelink-api/internal/model"
)

type ReqProduct struct {
	Id         int                       `json:"id"`
	Identifier string                    `json:"identifier" binding:"required"`
	Name       string                    `json:"name" binding:"required"`
	ModelId    int                       `json:"model_id" binding:"required"`
	Protocol   model.ProductProtocolType `json:"protocol" binding:"required"` // 产品使用的协议
}
