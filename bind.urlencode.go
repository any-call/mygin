package mygin

import (
	"net/http"
	"net/url"
)

type FormUrlEncode struct {
}

func (FormUrlEncode) Name() string {
	return "x-www-form-urlencoded"
}

func (FormUrlEncode) Bind(req *http.Request, obj any) error {
	return nil
}

func (FormUrlEncode) BindBody(body []byte, obj any) error {
	values, err := url.ParseQuery(string(body))
	if err != nil {
		return err
	}

	if err = mapForm(obj, values); err != nil {
		return err
	}

	return nil
}
