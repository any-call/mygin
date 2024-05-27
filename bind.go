package mygin

import (
	"encoding/json"
	"fmt"
	"github.com/any-call/gobase/util/myvalidator"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"net/http"
	"net/url"
	"reflect"
)

func SetServerError(code int, httpCode int) {
	serverError = code
	serverErrorHttpCode = httpCode
}

func SetBindParamError(code int, httpCode int) {
	bindError = code
	bindErrorHttpCode = httpCode
}

func Query(ctx *gin.Context, thenFunc noReqNoRespThenFunc) {
	do[noReq, noResp](ctx, noReq{}, nil, nil, nil, noReqNoRespThenFuncWrap(thenFunc))
}

func QueryReq[REQ any](ctx *gin.Context, req REQ, thenFunc reqNoRespThenFunc[REQ]) {
	do[REQ, noResp](ctx, req, bindQuery[REQ], validate[REQ], check[REQ], reqNoRespThenFuncWrap[REQ](thenFunc))
}

func QueryResp[RESP any](ctx *gin.Context, thenFunc noReqRespThenFunc[RESP]) {
	do[noReq, RESP](ctx, noReq{}, nil, nil, nil, noReqRespThenFuncWrap[RESP](thenFunc))
}

func QueryReqResp[REQ, RESP any](ctx *gin.Context, req REQ, thenFunc thenFunc[REQ, RESP]) {
	do[REQ, RESP](ctx, req, bindQuery[REQ], validate[REQ], check[REQ], thenFunc)
}

func Body(ctx *gin.Context, thenFunc noReqNoRespThenFunc) {
	do[noReq, noResp](ctx, noReq{}, nil, nil, nil, noReqNoRespThenFuncWrap(thenFunc))
}

func BodyReq[REQ any](ctx *gin.Context, req REQ, thenFunc reqNoRespThenFunc[REQ]) {
	do[REQ, noResp](ctx, req, bindJSON[REQ], validate[REQ], check[REQ], reqNoRespThenFuncWrap[REQ](thenFunc))
}

func BodyResp[RESP any](ctx *gin.Context, thenFunc noReqRespThenFunc[RESP]) {
	do[noReq, RESP](ctx, noReq{}, nil, nil, nil, noReqRespThenFuncWrap[RESP](thenFunc))
}

func BodyReqResp[REQ, RESP any](ctx *gin.Context, req REQ, thenFunc thenFunc[REQ, RESP]) {
	do[REQ, RESP](ctx, req, bindJSON[REQ], validate[REQ], check[REQ], thenFunc)
}

func Form(ctx *gin.Context, thenFunc noReqNoRespThenFunc) {
	do[noReq, noResp](ctx, noReq{}, nil, nil, nil, noReqNoRespThenFuncWrap(thenFunc))
}

func FormReq[REQ any](ctx *gin.Context, req REQ, thenFunc reqNoRespThenFunc[REQ]) {
	do[REQ, noResp](ctx, req, bindForm[REQ], validate[REQ], check[REQ], reqNoRespThenFuncWrap[REQ](thenFunc))
}

func FormResp[RESP any](ctx *gin.Context, thenFunc noReqRespThenFunc[RESP]) {
	do[noReq, RESP](ctx, noReq{}, nil, nil, nil, noReqRespThenFuncWrap[RESP](thenFunc))
}

func FormReqResp[REQ, RESP any](ctx *gin.Context, req REQ, thenFunc thenFunc[REQ, RESP]) {
	do[REQ, RESP](ctx, req, bindForm[REQ], validate[REQ], check[REQ], thenFunc)
}

func UriEncode(ctx *gin.Context, thenFunc noReqNoRespThenFunc) {
	do[noReq, noResp](ctx, noReq{}, nil, nil, nil, noReqNoRespThenFuncWrap(thenFunc))
}

func UriEncodeReq[REQ any](ctx *gin.Context, req REQ, thenFunc reqNoRespThenFunc[REQ]) {
	do[REQ, noResp](ctx, req, bindFormUriEncode[REQ], validate[REQ], check[REQ], reqNoRespThenFuncWrap[REQ](thenFunc))
}

func UriEncodeResp[RESP any](ctx *gin.Context, thenFunc noReqRespThenFunc[RESP]) {
	do[noReq, RESP](ctx, noReq{}, nil, nil, nil, noReqRespThenFuncWrap[RESP](thenFunc))
}

func UriEncodeReqResp[REQ, RESP any](ctx *gin.Context, req REQ, thenFunc thenFunc[REQ, RESP]) {
	do[REQ, RESP](ctx, req, bindFormUriEncode[REQ], validate[REQ], check[REQ], thenFunc)
}

func do[REQ, RESP any](ctx *gin.Context, req REQ, bindFunc bindFunc[REQ], validateFunc validateFunc[REQ], checkFunc checkFunc[REQ], thenFunc thenFunc[REQ, RESP]) {
	if fn := bindFunc; fn != nil {
		if err := fn(ctx, &req); err != nil {
			WriteBindError(ctx, err)
			return
		}
	}
	if fn := validateFunc; fn != nil {
		if err := fn(&req); err != nil {
			WriteBindError(ctx, err)
			return
		}
	}
	if fn := checkFunc; fn != nil {
		if err := fn(&req); err != nil {
			WriteBindError(ctx, err)
			return
		}
	}

	if fn := thenFunc; fn != nil {
		if resp, log, err := thenFunc(req); err != nil {
			ctx.Set("result", err.Error())
			if log != nil {
				ctx.Set("logs", log)
			}

			WriteServerErrorJSON(ctx, err)
		} else {
			ctx.Set("result", "")
			if log != nil {
				ctx.Set("logs", log)
			}

			if respType := reflect.TypeOf(resp); respType != nil {
				respTypeKind := respType.Kind()
				switch {
				case respTypeKind == reflect.String: // for func() (str string, err error)
					respStr := fmt.Sprintf("%v", resp)
					ctx.String(http.StatusOK, respStr)
				default:
					{
						WriteSuccessJSON(ctx, resp)
					}

				}
			}
		}
	}
}

type (
	noReq  struct{}
	noResp struct{}

	bindFunc[REQ any]       func(ctx *gin.Context, req *REQ) (err error)
	validateFunc[REQ any]   func(req *REQ) (err error)
	checkFunc[REQ any]      func(req *REQ) (err error)
	thenFunc[REQ, RESP any] func(req REQ) (resp RESP, log any, err error)

	reqNoRespThenFunc[REQ any]  func(req REQ) (log any, err error)
	noReqRespThenFunc[RESP any] func() (resp RESP, log any, err error)
	noReqNoRespThenFunc         func() (log any, err error)
)

func reqNoRespThenFuncWrap[REQ any](thenFunc reqNoRespThenFunc[REQ]) thenFunc[REQ, noResp] {
	return func(req REQ) (resp noResp, log any, err error) { log, err = thenFunc(req); return }
}

func noReqRespThenFuncWrap[RESP any](thenFunc noReqRespThenFunc[RESP]) thenFunc[noReq, RESP] {
	return func(req noReq) (resp RESP, log any, err error) { resp, log, err = thenFunc(); return }
}

func noReqNoRespThenFuncWrap(thenFunc noReqNoRespThenFunc) thenFunc[noReq, noResp] {
	return func(req noReq) (resp noResp, log any, err error) { log, err = thenFunc(); return }
}

func haveEncryptionData(ctx *gin.Context, dataType string) bool {
	b1 := false
	b2 := false
	{
		if value, ok := ctx.Get("have_encryption_data"); ok {
			if data, okk := value.(string); okk {
				if data == "yes" {
					b1 = true
				}
			}
		}
	}
	{
		if value, ok := ctx.Get("encryption_data_type"); ok {
			if data, okk := value.(string); okk {
				if data == dataType {
					b2 = true
				}
			}
		}
	}
	return b1 && b2
}

func bindQuery[REQ any](ctx *gin.Context, req *REQ) (err error) {
	if haveEncryptionData(ctx, "query") {
		if value, ok := ctx.Get("encryption_data"); ok {
			if values, okk := value.(url.Values); okk {
				return mapForm(req, values)
			}
		}
	}
	return ctx.ShouldBindQuery(req)
}

func bindJSON[REQ any](ctx *gin.Context, req *REQ) (err error) {
	if haveEncryptionData(ctx, "body") {
		if value, ok := ctx.Get("encryption_data"); ok {
			if values, okk := value.([]byte); okk {
				return json.Unmarshal(values, req)
			}
		}
	}
	return ctx.ShouldBindJSON(req)
}

func bindForm[REQ any](ctx *gin.Context, req *REQ) (err error) {
	if haveEncryptionData(ctx, "body") {
		if value, ok := ctx.Get("encryption_data"); ok {
			if values, okk := value.([]byte); okk {
				return json.Unmarshal(values, req)
			}
		}
	}
	return ctx.ShouldBindWith(req, binding.Form)
}

func bindFormUriEncode[REQ any](ctx *gin.Context, req *REQ) (err error) {
	if haveEncryptionData(ctx, "body") {
		if value, ok := ctx.Get("encryption_data"); ok {
			if values, okk := value.([]byte); okk {
				return json.Unmarshal(values, req)
			}
		}
	}
	return ctx.ShouldBindBodyWith(req, formUrlEncode{})
}

func validate[REQ any](req *REQ) (err error) {
	err = myvalidator.Validate(req)
	if err != nil {
		if trsMsg := translateMsg(req, err.Error()); trsMsg != nil {
			return trsMsg
		}
	}

	return err
}

func check[REQ any](req *REQ) (err error) {
loop:
	for _, value := range []reflect.Value{
		reflect.ValueOf(req).MethodByName("Check"),
		reflect.ValueOf(&req).MethodByName("Check"),
	} {
		if value.IsValid() {
			if values := value.Call([]reflect.Value{}); values != nil && len(values) > 0 {
				for _, val := range values {
					if val.CanInterface() {
						if inter := val.Interface(); inter != nil {
							if err = inter.(error); err != nil {
								break loop
							}
						}
					}
				}
			}
		}
	}
	return
}

func translateMsg[REQ any](req *REQ, key string) (err error) {
loop:
	for _, value := range []reflect.Value{
		reflect.ValueOf(req).MethodByName("TranslateLanguage"),
		reflect.ValueOf(&req).MethodByName("TranslateLanguage"),
	} {
		if value.IsValid() {
			if values := value.Call([]reflect.Value{reflect.ValueOf(key)}); values != nil && len(values) > 0 {
				for _, val := range values {
					if val.CanInterface() {
						if inter := val.Interface(); inter != nil {
							if err = inter.(error); err != nil {
								break loop
							}
						}
					}
				}
			}
		}
	}
	return
}
