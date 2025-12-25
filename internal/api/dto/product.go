package dto

import (
	"edgelink-api/internal/model"
)

type ReqProduct struct {
	Id         int                       `json:"id"`
	Identifier string                    `json:"identifier"`
	Name       string                    `json:"name"`
	ModelId    int                       `json:"model_id"`
	Protocol   model.ProductProtocolType `json:"protocol"` // 产品使用的协议
}
