package mygin

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type kv struct {
	Code int    `json:"code"`
	Msg  string `json:"msg,omitempty"`
	Data any    `json:"data,omitempty"`
}

func WriteJSON(ctx *gin.Context, code, httpCode int, msg string, err error, data any) {
	dd := kv{Code: code, Msg: msg, Data: data}
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
	WriteJSON(ctx, 0, http.StatusOK, "", nil, data)
}

func WriteServerErrorJSON(ctx *gin.Context, err error) {
	WriteJSON(ctx, http.StatusInternalServerError, http.StatusOK, "error", err, nil)
}

func WriteBindError(ctx *gin.Context, err error) {
	WriteJSON(ctx, http.StatusBadRequest, http.StatusOK, "", err, nil)
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
