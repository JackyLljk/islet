package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ResponseData struct {
	Code    MyCode `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func ResponseErrorWithMsg(ctx *gin.Context, code MyCode, errMsg string) {
	rd := &ResponseData{
		Code:    code,
		Message: errMsg,
		Data:    nil,
	}
	ctx.JSON(http.StatusOK, rd)
}

func ResponseSuccess(ctx *gin.Context, data any) {
	rd := &ResponseData{
		Code:    CodeSuccess,
		Message: CodeSuccess.Msg(),
		Data:    data,
	}
	ctx.JSON(http.StatusOK, rd)
}

func ResponseError(ctx *gin.Context, mc MyCode) {
	rd := &ResponseData{
		Code:    mc,
		Message: mc.Msg(),
		Data:    nil,
	}
	ctx.JSON(http.StatusOK, rd)
}
