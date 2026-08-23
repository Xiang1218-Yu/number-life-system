package service

import "testing"

func TestV6PaginationIncludesPartialPage(t *testing.T) {
	if got := TotalPages(21, 20); got != 2 {
		panic("pagination total pages mismatch")
	}
}
