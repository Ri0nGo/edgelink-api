package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code RespCode `json:"code"`
	Msg  string   `json:"msg"`
	Data any      `json:"data"`
}

func jsonResponse(c *gin.Context, code RespCode, msg string, data any) {
	c.JSON(http.StatusOK, Response{
		Code: code,
		Msg:  msg,
		Data: data,
	})
}

func Success(ctx *gin.Context, data ...any) {
	if len(data) > 0 {
		jsonResponse(ctx, RespCodeSuccess, RespCodeSuccess.msg(), data[0])
	} else {
		jsonResponse(ctx, RespCodeSuccess, RespCodeSuccess.msg(), nil)
	}
}

// Failed 响应错误，msg有值则用，无值则使用code的错误内容
func Failed(ctx *gin.Context, code RespCode, msg ...string) {
	if len(msg) > 0 {
		jsonResponse(ctx, code, msg[0], nil)
	} else {
		jsonResponse(ctx, code, code.msg(), nil)
	}
}
