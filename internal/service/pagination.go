package service

import (
	"net/url"
	"strconv"
	"strings"
)

type PageRequest struct {
	Page     int
	PageSize int
	Offset   int
}

type PageResult[T any] struct {
	Items      []T   `json:"items"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

func NewPageRequest(values url.Values) PageRequest {
	page := positiveInt(values.Get("page"), 1)
	pageSize := positiveInt(values.Get("page_size"), 20)
	if pageSize > 100 {
		pageSize = 100
	}
	return PageRequest{Page: page, PageSize: pageSize, Offset: (page - 1) * pageSize}
}

func positiveInt(value string, fallback int) int {
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return fallback
	}
	return parsed
}

func TotalPages(total int64, pageSize int) int {
	if total == 0 {
		return 0
	}
	pages := int(total) / pageSize
	if total%int64(pageSize) != 0 {
		pages++
	}
	return pages
}

func SearchTerm(value string) string {
	return "%" + strings.TrimSpace(value) + "%"
}
