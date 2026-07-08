package mygin

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
)

type (
	ID struct {
		ID int `json:"id" form:"id" validate:"min(1,入参ID不能为空)"`
	}

	IDs struct {
		IDs []int `json:"ids" form:"ids" validate:"minlength(1,入参ID不能为空)"`
	}

	PageReq struct {
		Limit int `form:"limit" json:"limit" validate:"min(1,入参limit不能为0)"`
		Page  int `form:"page" json:"page" validate:"min(1,入参page不能为0)"`
	}
	PageResp[T any] struct {
		Total int64 `json:"total"`
		Page  int   `json:"page"`
		Limit int   `json:"limit"`
		List  []T   `json:"list"`
	}

	BaseResp[T any] struct {
		Code int    `json:"code"`
		Msg  string `json:"msg,omitempty"`
		Data T      `json:"data,omitempty"`
	}

	PaginationOption func(*paginationConfig)

	paginationConfig struct {
		countDB  *gorm.DB
		count    int64
		hasCount bool
	}
)

func WithPaginationCountDB(db *gorm.DB) PaginationOption {
	return func(cfg *paginationConfig) {
		cfg.countDB = db
	}
}

func WithPaginationCount(count int64) PaginationOption {
	return func(cfg *paginationConfig) {
		cfg.count = count
		cfg.hasCount = true
	}
}

func SetServerError(code int, httpCode int) {
	serverError = code
	serverErrorHttpCode = httpCode
}

func SetBindParamError(code int, httpCode int) {
	bindError = code
	bindErrorHttpCode = httpCode
}

func SetRespCode(code int) {
	respSuccessCode = code
}

func (p PageResp[T]) TotalPage() int {
	if p.Limit <= 0 {
		return 0 // 防止除以零的情况
	}

	totalPages := (p.Total + int64(p.Limit) - 1) / int64(p.Limit) // 向上取整计算页数
	return int(totalPages)
}

func (self PageReq) Offset() int {
	myPage := self.Page
	if self.Page <= 0 {
		myPage = 1
	}

	return self.Limit * (myPage - 1)
}

func (self BaseResp[T]) Error() error {
	if self.Code != respSuccessCode {
		return fmt.Errorf(self.Msg)
	}

	return nil
}

func PaginationBySQL[T any](db *gorm.DB, baseSql string, req PageReq, resp *PageResp[T]) (err error) {
	total, list, page, limit, err := PaginationBySQLLimitPage[T](db, baseSql, req.Limit, req.Page)
	if err != nil {
		return err
	}

	resp.Total = total
	resp.Page = page
	resp.Limit = limit
	resp.List = list
	return nil
}

func Pagination[T any](db *gorm.DB, req PageReq, resp *PageResp[T]) (err error) {
	total, list, page, limit, err := PaginationByLimitPage[T](db, req.Limit, req.Page)
	if err != nil {
		return err
	}

	resp.Total = total
	resp.Page = page
	resp.Limit = limit
	resp.List = list
	return nil
}

func PaginationUsingCount[T any](db *gorm.DB, req PageReq, resp *PageResp[T], count int64) (err error) {
	total, list, page, limit, err := PaginationByLimitPage[T](db, req.Limit, req.Page, WithPaginationCount(count))
	if err != nil {
		return err
	}

	resp.Total = total
	resp.Page = page
	resp.Limit = limit
	resp.List = list
	return nil
}

//func Pagination_other(db *gorm.DB, limit, page int, count *int64, list any) (err error) {
//	if err = db.Count(count).Error; err != nil {
//		return
//	}
//	if *count == 0 {
//		return
//	}
//
//	if limit <= 0 {
//		limit = 10
//	}
//
//	if page <= 0 {
//		page = 1
//	}
//
//	err = db.Offset(limit * (page - 1)).Limit(limit).Find(list).Error
//
//	return
//}

func PaginationByLimitPage[T any](db *gorm.DB, limit, page int, opts ...PaginationOption) (total int64, list []T, realPage int, realLimit int, err error) {
	realLimit, realPage = normalizePageLimit(limit, page)
	cfg := newPaginationConfig(opts...)

	if total, err = paginationCount(db, cfg, func() (int64, error) {
		var total int64
		err := db.Count(&total).Error
		return total, err
	}); err != nil {
		return
	}

	if total == 0 {
		return
	}

	err = db.Offset(realLimit * (realPage - 1)).Limit(realLimit).Find(&list).Error
	return
}

func PaginationBySQLLimitPage[T any](db *gorm.DB, baseSql string, limit, page int, opts ...PaginationOption) (total int64, list []T, realPage int, realLimit int, err error) {
	realLimit, realPage = normalizePageLimit(limit, page)
	cfg := newPaginationConfig(opts...)

	if total, err = paginationCount(db, cfg, func() (int64, error) {
		var total int64
		countSql := fmt.Sprintf("SELECT COUNT(*) FROM (%s) AS t", baseSql)
		err := db.Raw(countSql).Scan(&total).Error
		return total, err
	}); err != nil {
		return
	}

	if total == 0 {
		return
	}

	pageSql := fmt.Sprintf("%s LIMIT ? OFFSET ?", baseSql)
	err = db.Raw(pageSql, realLimit, realLimit*(realPage-1)).Scan(&list).Error
	return
}

func normalizePageLimit(limit, page int) (realLimit, realPage int) {
	if limit <= 0 {
		limit = 10
	}

	if page <= 0 {
		page = 1
	}

	return limit, page
}

func newPaginationConfig(opts ...PaginationOption) paginationConfig {
	cfg := paginationConfig{}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	return cfg
}

func paginationCount(db *gorm.DB, cfg paginationConfig, defaultCount func() (int64, error)) (int64, error) {
	if cfg.hasCount {
		return cfg.count, nil
	}

	if cfg.countDB != nil {
		var total int64
		return total, cfg.countDB.Count(&total).Error
	}

	return defaultCount()
}

func PaginationFromArray[T any](list []T, req PageReq) (*PageResp[T], error) {
	if len(list) == 0 {
		return nil, fmt.Errorf("list is nil")
	}

	if req.Limit <= 0 {
		req.Limit = 10
	}

	if req.Page <= 0 {
		req.Page = 1
	}

	start := (req.Page - 1) * req.Limit
	end := start + req.Limit
	if end > len(list) {
		end = len(list)
	}

	if start > end {
		start = end
	}

	// 可能需要根据实际情况提供正确的 Total
	total := int64(len(list)) // 这里只是示例，实际总数应根据具体情况计算
	return &PageResp[T]{
		Total: total,
		Page:  req.Page,
		Limit: req.Limit,
		List:  list[start:end],
	}, nil
}

func GetOriginIP(ctx *gin.Context) string {
	rmtIP := ctx.GetHeader("X-Real-IP")
	if len(rmtIP) > 0 {
		return rmtIP
	}

	return ctx.RemoteIP()
}

func ParseResponse[DATA any](ret []byte, httpCode int, onCodeErr func(code int)) (info DATA, err error) {
	if httpCode == http.StatusOK {
		tmp := BaseResp[DATA]{}
		if err = json.Unmarshal(ret, &tmp); err != nil {
			return
		}

		if err = tmp.Error(); err != nil {
			if onCodeErr != nil {
				onCodeErr(tmp.Code)
			}
			return
		}

		info = tmp.Data
		return
	} else {
		if onCodeErr != nil {
			onCodeErr(httpCode)
		}
	}

	err = fmt.Errorf("%s", string(ret))
	return
}
