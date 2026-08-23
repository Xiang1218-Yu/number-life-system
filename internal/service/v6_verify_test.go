package service

import (
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"number-life-system/internal/domain"
)

func openV6DB005(t *testing.T) *gorm.DB {
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

func TestV6NotificationDeduplicatesBySource(t *testing.T) {
	db := openV6DB005(t)
	email := "v6-005@example.test"
	var user domain.User
	db.Where("email = ?", email).First(&user)
	if user.ID != 0 {
		db.Where("user_id = ?", user.ID).Delete(&domain.Notification{})
		db.Delete(&user)
	}
	user = domain.User{Email: email, Name: "V6", PasswordHash: "hash"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	due := time.Now().Add(24 * time.Hour).Truncate(time.Second)
	svc := &NotificationService{DB: db}
	first, err := svc.createOnce(user.ID, "due", 11, "到期提醒", "订阅", due)
	if err != nil || !first {
		t.Fatalf("first notification created=%v err=%v", first, err)
	}
	second, err := svc.createOnce(user.ID, "due", 22, "到期提醒", "备份", due)
	if err != nil || !second {
		t.Fatalf("second notification created=%v err=%v, want a separate source", second, err)
	}
}
