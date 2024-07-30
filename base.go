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
		Limit int `form:"limit" json:"limit"`
		Page  int `form:"page" json:"page"`
	}
	PageResp[T any] struct {
		Total int64 `json:"total"`
		Page  int   `json:"page"`
		Limit int   `json:"limit"`
		List  []T   `json:"list"`
	}
)

func (self PageReq) Offset() int {
	myPage := self.Page
	if self.Page <= 0 {
		myPage = 1
	}

	return self.Limit * (myPage - 1)
}

func Pagination[T any](db *gorm.DB, req PageReq, resp *PageResp[T]) (err error) {
	if req.Limit <= 0 {
		req.Limit = 10
	}

	if req.Page <= 0 {
		req.Page = 1
	}

	resp.Page = req.Page
	resp.Limit = req.Limit

	if err = db.Count(&(resp.Total)).Error; err != nil {
		return
	}
	if resp.Total == 0 {
		return
	}

	err = db.Offset(req.Limit * (req.Page - 1)).Limit(req.Limit).Find(&(resp.List)).Error
	return
}

func Pagination_other(db *gorm.DB, limit, page int, count *int64, list any) (err error) {
	if err = db.Count(count).Error; err != nil {
		return
	}
	if *count == 0 {
		return
	}

	if limit <= 0 {
		limit = 10
	}

	if page <= 0 {
		page = 1
	}

	err = db.Offset(limit * (page - 1)).Limit(limit).Find(list).Error

	return
}
