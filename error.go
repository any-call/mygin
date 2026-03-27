package mygin

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Error struct {
	error          string
	httpCode, code int
}

func NewError(error string) *Error { return &Error{error: error} }

func NewErrorWithCode(error string, code int) *Error {
	return &Error{error: error, code: code}
}

func NewErrorWithHttpCode(error string, httpCode int) *Error {
	return &Error{error: error, httpCode: httpCode}
}

func NewErrorWithCodes(error string, httpCode int, code int) *Error {
	return &Error{error: error, httpCode: httpCode, code: code}
}

func (e *Error) Error() string { return e.error }

// / 自定义的错误处理
type WriteError interface {
	error
	WriteTo(ctx *gin.Context)
}

func tryWriteError(ctx *gin.Context, err error) bool {
	if err == nil {
		return false
	}

	var target WriteError
	if errors.As(err, &target) {
		target.WriteTo(ctx)
		return true
	}

	return false
}

type writeError struct {
	err   error
	write func(ctx *gin.Context)
}

func (e *writeError) Error() string {
	if e.err != nil {
		return e.err.Error()
	}
	return ""
}

func (e *writeError) WriteTo(ctx *gin.Context) {
	if e.write != nil {
		e.write(ctx)
	}
}

func NewWriteError(err error, fn func(ctx *gin.Context)) error {
	return &writeError{
		err:   err,
		write: fn,
	}
}

func NewStringError(httpCode int, body string) error {
	return &writeError{
		err: errors.New(body),
		write: func(ctx *gin.Context) {
			ctx.String(httpCode, body)
		},
	}
}

func NewStatusError(httpCode int) error {
	return &writeError{
		err: errors.New(http.StatusText(httpCode)),
		write: func(ctx *gin.Context) {
			ctx.Status(httpCode)
		},
	}
}

func NewJSONError(code int, msg string) error {
	return &writeError{
		err: NewErrorWithCodes(msg, http.StatusOK, code),
		write: func(ctx *gin.Context) {
			WriteJSON(ctx, code, http.StatusOK, msg, nil, nil)
		},
	}
}
