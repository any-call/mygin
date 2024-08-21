package mygin

import "testing"

func TestPageResp_TotalPage(t *testing.T) {
	pageResp := &PageResp[int]{
		Total: 21,
		Limit: 20,
	}

	t.Log("total page:", pageResp.TotalPage())
}
