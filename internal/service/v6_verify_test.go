package service

import (
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"number-life-system/internal/domain"
)

func openV6DB004(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(postgres.Open("host=localhost user=postgres password=postgres dbname=digital_life port=5432 sslmode=disable TimeZone=Asia/Shanghai"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&domain.User{}, &domain.AccountCategory{}, &domain.Account{}, &domain.Subscription{}, &domain.DigitalFootprint{}, &domain.DataLocation{}, &domain.BackupRecord{}, &domain.Notification{}, &domain.SecurityEvent{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestV6CSVImportPreservesArchivedStatus(t *testing.T) {
	db := openV6DB004(t)
	email := "v6-004@example.test"
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
	csvData := []byte("platform,username,email,category,registered_at,password_strength,password_changed_at,two_factor_enabled,known_breach,last_login_at,notes,status\nExample,v6,,other,,strong,,false,false,,,archived\n")
	if _, err := (&CSVService{DB: db}).ImportAccounts(user.ID, csvData); err != nil {
		t.Fatal(err)
	}
	var row domain.Account
	if err := db.Where("user_id = ? AND platform = ?", user.ID, "Example").First(&row).Error; err != nil {
		t.Fatal(err)
	}
	if row.Status != "archived" {
		t.Fatalf("imported status = %q, want archived", row.Status)
	}
}
