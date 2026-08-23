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

// HasNext reports whether a subsequent page exists. TotalPages is authoritative
// because it accounts for the trailing partial page; comparing Total against
// the current offset would under-count the same way the buggy divisor once did.
func (p PageResult[T]) HasNext() bool { return p.Page < p.TotalPages }

// HasPrev reports whether an earlier page exists.
func (p PageResult[T]) HasPrev() bool { return p.Page > 1 }

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

// TotalPages returns the number of pages needed to hold `total` items at the
// given `pageSize`. The trailing partial page counts as a full page: e.g. 105
// items at page size 20 → 6 pages (five full pages plus one holding the last 5).
// A non-positive page size is treated as a single page so the navigation stays
// well-defined instead of dividing by zero.
func TotalPages(total int64, pageSize int) int {
	if total <= 0 {
		return 0
	}
	if pageSize <= 0 {
		return 1
	}
	pages := int(total / int64(pageSize))
	if total%int64(pageSize) != 0 {
		pages++
	}
	return pages
}

func SearchTerm(value string) string {
	return "%" + strings.TrimSpace(value) + "%"
}
