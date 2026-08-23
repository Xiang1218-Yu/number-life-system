package service

import (
	"fmt"
	"os"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"number-life-system/internal/domain"
)

// This is a real end-to-end regression test for the pagination off-by-one bug.
// It exercises the full pipeline — SQL Count, Limit/Offset, and the corrected
// TotalPages — against a live database, then walks every page to prove the
// trailing partial page is reachable and that has-next/has-prev are correct on
// the last page (the exact navigation entry the bug used to drop).
//
// It is skipped unless NLS_TEST_DSN points at a Postgres instance, so it is
// inert in environments without a database.
func TestListPagePaginationEndToEnd(t *testing.T) {
	dsn := os.Getenv("NLS_TEST_DSN")
	if dsn == "" {
		t.Skip("NLS_TEST_DSN not set; skipping live pagination regression test")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&domain.User{}, &domain.AccountCategory{}, &domain.Account{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	// Unique per-run identifier so concurrent runs and reruns do not collide.
	run := fmt.Sprintf("pgtest-%d", time.Now().UnixNano())
	user := domain.User{Email: run + "@test.local", PasswordHash: "x", Name: "Tester"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	t.Cleanup(func() {
		db.Where("user_id = ?", user.ID).Delete(&domain.Account{})
		db.Delete(&user)
	})

	const total = 55
	const pageSize = 20
	rows := make([]domain.Account, 0, total)
	for i := 0; i < total; i++ {
		rows = append(rows, domain.Account{
			UserID:           user.ID,
			Platform:         fmt.Sprintf("%s-platform-%d", run, i),
			Username:         fmt.Sprintf("%s-user-%d", run, i),
			Category:         "development",
			PasswordStrength: "strong",
			Status:           "active",
		})
	}
	if err := db.Create(&rows).Error; err != nil {
		t.Fatalf("seed accounts: %v", err)
	}

	svc := &AccountService{DB: db}
	const wantPages = 3 // 55 / 20 → 20 + 20 + 15

	// Walk every page and assert the navigation flags are correct at each step,
	// including (critically) the last partial page the old divisor dropped.
	for page := 1; page <= wantPages; page++ {
		filter := AccountFilter{Page: PageRequest{Page: page, PageSize: pageSize, Offset: (page - 1) * pageSize}}
		result, err := svc.ListPage(user.ID, false, filter)
		if err != nil {
			t.Fatalf("ListPage(%d): %v", page, err)
		}
		if result.Total != total {
			t.Fatalf("page %d total = %d, want %d", page, result.Total, total)
		}
		if result.TotalPages != wantPages {
			t.Fatalf("page %d total_pages = %d, want %d (off-by-one regression)", page, result.TotalPages, wantPages)
		}
		wantItems := pageSize
		if page == wantPages {
			wantItems = total - (wantPages-1)*pageSize // trailing 15
		}
		if len(result.Items) != wantItems {
			t.Fatalf("page %d items = %d, want %d", page, len(result.Items), wantItems)
		}
		if got := result.HasPrev(); got != (page > 1) {
			t.Fatalf("page %d HasPrev = %v, want %v", page, got, page > 1)
		}
		if got := result.HasNext(); got != (page < wantPages) {
			t.Fatalf("page %d HasNext = %v, want %v", page, got, page < wantPages)
		}
	}

	// Requesting the last page directly must succeed — the bug returned a
	// total_pages short of reality, so the frontend never offered the entry.
	last, err := svc.ListPage(user.ID, false, AccountFilter{Page: PageRequest{Page: wantPages, PageSize: pageSize, Offset: (wantPages - 1) * pageSize}})
	if err != nil {
		t.Fatalf("load last page: %v", err)
	}
	if last.HasNext() {
		t.Fatalf("last page reports HasNext=true; navigation would loop")
	}
	if len(last.Items) == 0 {
		t.Fatalf("last page is empty; the trailing page was dropped again")
	}
}
