package service

import (
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"number-life-system/internal/domain"
)

func openV6DB006(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:bug006?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)
	if err := db.AutoMigrate(&domain.User{}, &domain.AccountCategory{}, &domain.Account{}, &domain.Subscription{}, &domain.DigitalFootprint{}, &domain.DataLocation{}, &domain.BackupRecord{}, &domain.Notification{}, &domain.SecurityEvent{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestBug6RecentLoginScore(t *testing.T) {
	db := openV6DB006(t)
	email := "v6-006@example.test"
	var user domain.User
	db.Where("email = ?", email).First(&user)
	if user.ID != 0 {
		db.Where("user_id = ?", user.ID).Delete(&domain.Account{})
		db.Delete(&user)
	}
	user = domain.User{Email: email, Name: "V6", PasswordHash: "hash"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	lastLogin := time.Now().Add(-24 * time.Hour)
	row := domain.Account{UserID: user.ID, Platform: "Example", Username: "v6", Category: "other", PasswordStrength: "strong", TwoFactorEnabled: true, LastLoginAt: &lastLogin, Status: "active"}
	if err := db.Create(&row).Error; err != nil {
		t.Fatal(err)
	}
	report, err := (&SecurityService{DB: db}).Report(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Accounts) != 1 || report.Accounts[0].Score != 100 {
		t.Fatalf("security score = %#v, want one account with score 100", report.Accounts)
	}
}
