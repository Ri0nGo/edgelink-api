package handler

import (
	err2 "edgelink-api/internal/pkg/bizerr"
	"edgelink-api/internal/pkg/ginx/response"
	"errors"

	"github.com/gin-gonic/gin"
)

// HandlerError 如果是自定义错误，则msg中显示自定义错误，不是则显示code所对应的msg
func HandlerError(ctx *gin.Context, code response.RespCode, err error) {
	var bizErr *err2.BizError

	// 自定义错误
	if err != nil && errors.As(err, &bizErr) {
		response.Failed(ctx, code, err.Error())
		return
	}

	response.Failed(ctx, code)
}
