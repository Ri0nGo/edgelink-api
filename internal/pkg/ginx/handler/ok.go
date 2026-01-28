package handler

import (
	"edgelink-api/internal/pkg/ginx/response"

	"github.com/gin-gonic/gin"
)

func Success(ctx *gin.Context, data ...any) {
	response.Success(ctx, data...)
}
