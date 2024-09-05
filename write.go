package mygin

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

var (
	serverError         int = http.StatusInternalServerError
	serverErrorHttpCode int = http.StatusOK

	bindError         int = http.StatusBadRequest
	bindErrorHttpCode int = http.StatusOK
)

func WriteJSON(ctx *gin.Context, code, httpCode int, msg string, err error, data any) {
	dd := BaseResp{Code: code, Msg: msg, Data: data}
	if err != nil {
		dd.Msg = err.Error()
		if cusErr, ok := err.(*Error); ok {
			if cusErr.code > 0 {
				dd.Code = cusErr.code
			}
			if cusErr.httpCode > 0 {
				httpCode = cusErr.httpCode
			}
		}
	}
	ctx.JSON(httpCode, dd)
}

func WriteSuccessJSON(ctx *gin.Context, data any) {
	WriteJSON(ctx, 0, http.StatusOK, "success", nil, data)
}

func WriteServerErrorJSON(ctx *gin.Context, err error) {
	WriteJSON(ctx, serverError, serverErrorHttpCode, "error", err, nil)
}

func WriteBindError(ctx *gin.Context, err error) {
	WriteJSON(ctx, bindError, bindErrorHttpCode, "", err, nil)
}

func WriteBindCodeSimple(httpCode int, ctx *gin.Context, str string) {
	WriteJSON(ctx, httpCode, httpCode, str, nil, nil)
}

func WriteJSONSimple(ctx *gin.Context, err error) {
	if err != nil {
		WriteBindError(ctx, err)
	} else {
		WriteSuccessJSON(ctx, nil)
	}
}
