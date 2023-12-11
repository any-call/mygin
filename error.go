package mygin

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
