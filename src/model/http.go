package model

import (
	"math"
)

type HTTPResponse struct {
	Meta       Meta             `json:"metaData"`
	Message    ResponseMessage  `json:"message"`
	IsSuccess  bool             `json:"isSuccess"`
	Data       interface{}      `json:"data"`
	Pagination *PaginationParam `json:"pagination"`
}

type ResponseMessage struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type Meta struct {
	Time        string `json:"timestamp"`
	RequestID   string `json:"requestId"`
	TimeElapsed string `json:"timeElapsed"`
}

type PaginationParam struct {
	Limit          int64 `form:"limit" json:"limit"`
	Page           int64 `form:"page" json:"-"`
	Offset         int64 `json:"-"`
	CurrentPage    int64 `json:"currentPage"`
	TotalPage      int64 `json:"totalPage"`
	CurrentElement int64 `json:"currentElement"`
	TotalElement   int64 `json:"totalElement"`
}

func (pg *PaginationParam) SetDefaultPagination() {
	if pg.Limit == 0 {
		pg.Limit = 10
	}

	if pg.Page == 0 {
		pg.Page = 1
	}

	pg.Offset = (pg.Page - 1) * pg.Limit
}

func (pg *PaginationParam) ProcessPagination(rowsAffected int64) {
	pg.CurrentPage = pg.Page
	pg.TotalPage = int64(math.Ceil(float64(pg.TotalElement) / float64(pg.Limit)))
	pg.CurrentElement = rowsAffected
}
