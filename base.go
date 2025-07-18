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
)

func SetServerError(code int, httpCode int) {
	serverError = code
	serverErrorHttpCode = httpCode
}

func SetBindParamError(code int, httpCode int) {
	bindError = code
	bindErrorHttpCode = httpCode
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
	if self.Code != 0 {
		return fmt.Errorf(self.Msg)
	}

	return nil
}

func PaginationBySQL[T any](db *gorm.DB, baseSql string, req PageReq, resp *PageResp[T]) (err error) {
	if req.Limit <= 0 {
		req.Limit = 10
	}

	if req.Page <= 0 {
		req.Page = 1
	}

	resp.Page = req.Page
	resp.Limit = req.Limit
	// 统计总数（包一层 count）
	countSql := fmt.Sprintf("SELECT COUNT(*) FROM (%s) AS t", baseSql)
	if err := db.Raw(countSql).Scan(&resp.Total).Error; err != nil {
		return err
	}

	if resp.Total == 0 {
		return
	}

	// 加分页参数
	offset := req.Limit * (req.Page - 1)
	pageSql := fmt.Sprintf("%s LIMIT ? OFFSET ?", baseSql)
	if err := db.Raw(pageSql, req.Limit, offset).Scan(&resp.List).Error; err != nil {
		return err
	}
	return
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

func PaginationUsingCount[T any](db *gorm.DB, req PageReq, resp *PageResp[T], count int64) (err error) {
	if req.Limit <= 0 {
		req.Limit = 10
	}

	if req.Page <= 0 {
		req.Page = 1
	}

	resp.Page = req.Page
	resp.Limit = req.Limit

	resp.Total = count
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
