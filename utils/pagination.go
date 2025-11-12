package utils

import (
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Pagination struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	TotalRows  int64 `json:"total_rows"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

func GetPagination(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))

	if page < 1 {
		page = 1
	}

	if perPage < 1 || perPage > 100 {
		perPage = 10
	}

	return page, perPage
}

func CalculatePagination(page, perPage int, totalRows int64) Pagination {
	totalPages := int(math.Ceil(float64(totalRows)/float64(perPage)))

	if totalRows == 0 {
		totalPages = 0
	}

	hasNext := page < totalPages
	hasPrev := page > 1

	return Pagination{
		Page: page,
		PerPage: perPage,
		TotalRows: totalRows,
		TotalPages: totalPages,
		HasNext: hasNext,
		HasPrev: hasPrev,
	}
}

func GetOffset(page, perPage int) int {
	return (page - 1) * perPage
}

func IsValidPage(page, totalPages int) bool {
	if totalPages == 0 {
		return page == 1
	}

	return page > 0 && page <= totalPages
}