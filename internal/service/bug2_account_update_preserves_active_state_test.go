package service

import (
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"number-life-system/internal/domain"
)

func openBug2DB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:bug2-account-state?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&domain.User{}, &domain.Account{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestBug2AccountUpdatePreservesActiveState(t *testing.T) {
	db := openBug2DB(t)
	user := domain.User{Email: "bug2@example.test", Name: "Bug 2", PasswordHash: "hash"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}

	svc := &AccountService{DB: db}
	row, err := svc.Create(user.ID, AccountInput{
		Platform: "Example",
		Username: "bug2-user",
		Category: "other",
		Password: "StrongPass1!",
	})
	if err != nil {
		t.Fatal(err)
	}

	updated, err := svc.Update(user.ID, row.ID, AccountInput{
		Platform:    "Example",
		Username:    "bug2-user",
		Category:    "other",
		LastLoginAt: ptrBug2Time(time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)),
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != "active" {
		t.Fatalf("updated account status = %q, want active", updated.Status)
	}
}

func ptrBug2Time(value time.Time) *time.Time { return &value }
