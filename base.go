package mygin

import (
	"gorm.io/gorm"
)

type (
	ID struct {
		ID int `json:"id" form:"id" validate:"min(1,入参ID不能为空)"`
	}

	IDs struct {
		IDs []int `json:"ids" form:"ids" validate:"minlength(1,入参ID不能为空)"`
	}

	PageReq struct {
		Limit int `form:"limit" validate:"min(1,无效的分页数据)"`
		Page  int `form:"page"`
	}
	PageResp[T any] struct {
		Total int64 `json:"total"`
		Page  int   `json:"page"`
		Limit int   `json:"limit"`
		List  []T   `json:"list"`
	}
)

func Pagination(db *gorm.DB, req PageReq, count *int64, list any) (err error) {
	if err = db.Count(count).Error; err != nil {
		return
	}
	if *count == 0 {
		return
	}

	if req.Limit <= 0 {
		req.Limit = 10
	}

	if req.Page <= 0 {
		req.Page = 1
	}

	err = db.Offset(req.Limit * (req.Page - 1)).Limit(req.Limit).Find(list).Error

	return
}

func PaginationEx[T any](db *gorm.DB, req PageReq, resp *PageResp[T]) (err error) {
	if err = db.Count(&(resp.Total)).Error; err != nil {
		return
	}
	if resp.Total == 0 {
		return
	}

	if req.Limit <= 0 {
		req.Limit = 10
	}

	if req.Page <= 0 {
		req.Page = 1
	}

	resp.Page = req.Page
	resp.Limit = req.Limit
	err = db.Offset(req.Limit * (req.Page - 1)).Limit(req.Limit).Find(&(resp.List)).Error

	return
}
