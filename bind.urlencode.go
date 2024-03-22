package mygin

import (
	"net/http"
	"net/url"
)

var (
	FormUrlEncode = formUrlEncode{}
)

type formUrlEncode struct {
}

func (formUrlEncode) Name() string {
	return "x-www-form-urlencoded"
}

func (formUrlEncode) Bind(req *http.Request, obj any) error {
	return nil
}

func (formUrlEncode) BindBody(body []byte, obj any) error {
	values, err := url.ParseQuery(string(body))
	if err != nil {
		return err
	}

	if err = mapForm(obj, values); err != nil {
		return err
	}

	return nil
}
