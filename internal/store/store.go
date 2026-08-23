package store

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"number-life-system/config"
	"number-life-system/internal/domain"
)

func Open(cfg config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.Database.DSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	if err := db.AutoMigrate(&domain.User{}, &domain.AccountCategory{}, &domain.Account{}, &domain.Subscription{}, &domain.DigitalFootprint{}, &domain.DataLocation{}, &domain.BackupRecord{}, &domain.Notification{}, &domain.SecurityEvent{}); err != nil {
		return nil, err
	}
	if err := seedCategories(db); err != nil {
		return nil, err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_notifications_reference ON notifications (user_id, reference_id)").Error; err != nil {
		return nil, err
	}
	// 与 createOnce 的去重键 (user_id, type, reference_id) 对齐：先幂等清理历史重复行，再建唯一索引。
	// 保留每组最早一条，其余删除。这能修复旧逻辑下因 due_at 不同而残留的重复通知。
	if err := db.Exec(`DELETE FROM notifications a USING notifications b
WHERE a.user_id = b.user_id AND a.type = b.type AND a.reference_id = b.reference_id AND a.id > b.id`).Error; err != nil {
		return nil, err
	}
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_notifications_user_type_reference ON notifications (user_id, type, reference_id)").Error; err != nil {
		return nil, err
	}
	return db, nil
}
func seedCategories(db *gorm.DB) error {
	items := []domain.AccountCategory{{Name: "social", Color: "#ff7a59"}, {Name: "cloud", Color: "#4c8bf5"}, {Name: "development", Color: "#8b5cf6"}, {Name: "finance", Color: "#16a085"}, {Name: "subscription", Color: "#f59e0b"}, {Name: "work", Color: "#334155"}, {Name: "other", Color: "#64748b"}}
	for _, item := range items {
		if err := db.Where("name = ?", item.Name).FirstOrCreate(&item).Error; err != nil {
			return err
		}
	}
	return nil
}
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
