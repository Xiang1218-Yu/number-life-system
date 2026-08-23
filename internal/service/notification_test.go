package service

import (
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"number-life-system/internal/domain"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	// 每个测试用独立的内存库，避免跨用例数据污染。
	dsn := "file:" + t.Name() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&domain.Notification{}, &domain.Subscription{}, &domain.BackupRecord{}, &domain.Account{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

// 同一来源（订阅）内，两条记录 due_at 完全相同，应各自保留一条提醒。
func TestRefreshSubscriptions_DifferentRecordsSameDueBothKept(t *testing.T) {
	db := newTestDB(t)
	s := &NotificationService{DB: db}
	due := time.Now().UTC().Add(3 * 24 * time.Hour).Truncate(time.Second)
	// 两条不同订阅，next_billing_at 相同
	if err := db.Create(&domain.Subscription{UserID: 1, ServiceName: "A", Amount: 12, Currency: "CNY", Cycle: "month", Status: "active", NextBillingAt: &due}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&domain.Subscription{UserID: 1, ServiceName: "B", Amount: 30, Currency: "CNY", Cycle: "month", Status: "active", NextBillingAt: &due}).Error; err != nil {
		t.Fatal(err)
	}
	created, err := s.refreshSubscriptions(1)
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if created != 2 {
		t.Fatalf("expected 2 notifications created, got %d", created)
	}
	var count int64
	db.Model(&domain.Notification{}).Where("user_id = ? AND type = ?", 1, "subscription_due").Count(&count)
	if count != 2 {
		t.Fatalf("expected 2 persisted rows, got %d", count)
	}
}

// 订阅与备份同时到期（due_at 相同），应各自保留——即"不同来源的到期事项都能保留"。
func TestRefresh_SubscriptionAndBackupSameDueBothKept(t *testing.T) {
	db := newTestDB(t)
	s := &NotificationService{DB: db}
	due := time.Now().UTC().Add(3 * 24 * time.Hour).Truncate(time.Second)
	if err := db.Create(&domain.Subscription{UserID: 1, ServiceName: "Cloud", Amount: 20, Currency: "CNY", Cycle: "month", Status: "active", NextBillingAt: &due}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&domain.BackupRecord{UserID: 1, Platform: "NAS", Cycle: "month", NextBackupAt: &due}).Error; err != nil {
		t.Fatal(err)
	}
	created, err := s.Refresh(1)
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if created < 2 {
		t.Fatalf("expected at least 2 (sub+backup) notifications created, got %d", created)
	}
	var sub, bk int64
	db.Model(&domain.Notification{}).Where("user_id = ? AND type = ?", 1, "subscription_due").Count(&sub)
	db.Model(&domain.Notification{}).Where("user_id = ? AND type = ?", 1, "backup_due").Count(&bk)
	if sub != 1 || bk != 1 {
		t.Fatalf("expected 1 sub + 1 backup, got sub=%d backup=%d", sub, bk)
	}
}

// 反复刷新同一批到期事项，不应重复创建（幂等）。
func TestRefresh_IdempotentAcrossRefreshes(t *testing.T) {
	db := newTestDB(t)
	s := &NotificationService{DB: db}
	due := time.Now().UTC().Add(3 * 24 * time.Hour).Truncate(time.Second)
	if err := db.Create(&domain.Subscription{UserID: 1, ServiceName: "A", Amount: 12, Currency: "CNY", Cycle: "month", Status: "active", NextBillingAt: &due}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&domain.BackupRecord{UserID: 1, Platform: "NAS", Cycle: "month", NextBackupAt: &due}).Error; err != nil {
		t.Fatal(err)
	}
	first, err := s.Refresh(1)
	if err != nil {
		t.Fatal(err)
	}
	if first != 2 {
		t.Fatalf("first refresh should create 2, got %d", first)
	}
	second, err := s.Refresh(1)
	if err != nil {
		t.Fatal(err)
	}
	if second != 0 {
		t.Fatalf("second refresh should create 0 (idempotent), got %d", second)
	}
}

// 改期后刷新：同一条记录的提醒应更新 due_at 而非新增一条。
func TestRefresh_RescheduleUpdatesDueAt(t *testing.T) {
	db := newTestDB(t)
	s := &NotificationService{DB: db}
	due1 := time.Now().UTC().Add(3 * 24 * time.Hour).Truncate(time.Second)
	sub := domain.Subscription{UserID: 1, ServiceName: "A", Amount: 12, Currency: "CNY", Cycle: "month", Status: "active", NextBillingAt: &due1}
	if err := db.Create(&sub).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := s.refreshSubscriptions(1); err != nil {
		t.Fatal(err)
	}
	// 改期到另一个仍在窗口内的日期
	due2 := due1.AddDate(0, 0, 2)
	db.Model(&domain.Subscription{}).Where("id = ?", sub.ID).Update("next_billing_at", due2)

	created, err := s.refreshSubscriptions(1)
	if err != nil {
		t.Fatal(err)
	}
	if created != 1 {
		t.Fatalf("reschedule refresh should report 1 update, got %d", created)
	}
	var rows []domain.Notification
	db.Where("user_id = ? AND type = ?", 1, "subscription_due").Find(&rows)
	if len(rows) != 1 {
		t.Fatalf("expected exactly 1 row after reschedule (updated, not duplicated), got %d", len(rows))
	}
	if !rows[0].DueAt.Equal(due2) {
		t.Fatalf("due_at not updated: want %v got %v", due2, rows[0].DueAt)
	}
}
