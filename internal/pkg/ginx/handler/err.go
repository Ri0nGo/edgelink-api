package handler

import (
	err2 "edgelink-api/internal/pkg/bizerr"
	"edgelink-api/internal/pkg/ginx/response"
	"errors"

	"github.com/gin-gonic/gin"
)

// HandlerError 如果是业务错误，则msg中显示业务错误，不是则显示code所对应的msg
func HandlerError(ctx *gin.Context, code response.RespCode, err error) {
	var bizErr *err2.BizError

	// 业务错误
	if err != nil && errors.As(err, &bizErr) {
		response.Failed(ctx, code, err.Error())
		return
	}

	response.Failed(ctx, code)
}
