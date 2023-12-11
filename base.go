package mygin

import (
	"gorm.io/gorm"
)

type (
	PageReq struct {
		Limit int `form:"limit" validate:"min(1,无效的分页数据)"`
		Page  int `form:"page"`
	}
	PageResp[T any] struct {
		Total int64 `json:"total"`
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
	if req.Limit == 0 {
		req.Limit = 10
	}

	if req.Page == 0 {
		req.Page = 1
	}

	err = db.Offset(req.Limit * (req.Page - 1)).Limit(req.Limit).Find(list).Error

	return
}
