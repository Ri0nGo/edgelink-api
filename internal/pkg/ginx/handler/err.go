package handler

import (
	err2 "edgelink-api/internal/pkg/bizerr"
	"edgelink-api/internal/pkg/ginx/response"
	"errors"

	"github.com/gin-gonic/gin"
)

// HandlerError 如果是业务错误，则msg中显示业务错误，不是则显示code所对应的msg
func HandlerError(ctx *gin.Context, code response.RespCode, errs ...error) {
	var bizErr *err2.BizError

	// 业务错误
	if len(errs) > 0 && errors.As(errs[0], &bizErr) {
		response.Failed(ctx, code, errs[0].Error())
		return

	}

	response.Failed(ctx, code)
}
