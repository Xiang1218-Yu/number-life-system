package service

import "testing"

func TestTotalPages(t *testing.T) {
	cases := []struct {
		name     string
		total    int64
		pageSize int
		want     int
	}{
		{"empty yields zero pages", 0, 20, 0},
		{"negative total yields zero pages", -5, 20, 0},
		{"exact multiple keeps full count", 100, 20, 5},
		{"trailing partial page counts", 105, 20, 6},
		{"single item is one page", 1, 20, 1},
		{"one item shy of a full page still rounds up", 19, 20, 1},
		{"one item over a full page adds a page", 21, 20, 2},
		{"page size of one item per page", 1000, 1, 1000},
		{"non-positive page size falls back to one page", 50, 0, 1},
		{"large total does not overflow int", 9_999_999_999, 50, 199_999_999 + 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := TotalPages(tc.total, tc.pageSize); got != tc.want {
				t.Fatalf("TotalPages(%d, %d) = %d, want %d", tc.total, tc.pageSize, got, tc.want)
			}
		})
	}
}

func TestPageResultNavigation(t *testing.T) {
	// 105 items / page size 20 → 6 pages. Page 6 is the trailing partial page
	// that the old under-counting divisor dropped, which broke the last-page
	// navigation entry.
	result := PageResult[struct{}]{
		Items:      make([]struct{}, 5),
		Page:       6,
		PageSize:   20,
		Total:      105,
		TotalPages: TotalPages(105, 20),
	}
	if result.TotalPages != 6 {
		t.Fatalf("TotalPages = %d, want 6", result.TotalPages)
	}
	if result.HasNext() {
		t.Fatalf("HasNext on last page = true, want false")
	}
	if !result.HasPrev() {
		t.Fatalf("HasPrev on page 6 = false, want true")
	}

	first := PageResult[struct{}]{Page: 1, PageSize: 20, Total: 105, TotalPages: 6}
	if !first.HasNext() {
		t.Fatalf("HasNext on first page = false, want true")
	}
	if first.HasPrev() {
		t.Fatalf("HasPrev on first page = true, want false")
	}
}
