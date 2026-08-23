package service

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"number-life-system/internal/domain"
)

type NotificationService struct {
	DB *gorm.DB
}

type NotificationFilter struct {
	Status string
	Type   string
	Limit  int
}

type NotificationSummary struct {
	Total   int64 `json:"total"`
	Pending int64 `json:"pending"`
	Read    int64 `json:"read"`
}

func (s *NotificationService) List(userID uint, filter NotificationFilter) ([]domain.Notification, error) {
	var rows []domain.Notification
	query := s.DB.Where("user_id = ?", userID)
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}
	limit := filter.Limit
	if limit < 1 || limit > 100 {
		limit = 50
	}
	err := query.Order("due_at asc").Limit(limit).Find(&rows).Error
	return rows, err
}

func (s *NotificationService) Summary(userID uint) (NotificationSummary, error) {
	var total, pending, read int64
	query := s.DB.Model(&domain.Notification{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return NotificationSummary{}, err
	}
	if err := query.Where("status = ?", "pending").Count(&pending).Error; err != nil {
		return NotificationSummary{}, err
	}
	if err := query.Where("status = ?", "read").Count(&read).Error; err != nil {
		return NotificationSummary{}, err
	}
	return NotificationSummary{Total: total, Pending: pending, Read: read}, nil
}

func (s *NotificationService) MarkRead(userID, id uint) error {
	result := s.DB.Model(&domain.Notification{}).Where("user_id = ? AND id = ?", userID, id).Updates(map[string]any{"status": "read"})
	if result.RowsAffected == 0 {
		return errors.New("提醒不存在")
	}
	return result.Error
}

func (s *NotificationService) MarkAllRead(userID uint) error {
	return s.DB.Model(&domain.Notification{}).Where("user_id = ? AND status = ?", userID, "pending").Updates(map[string]any{"status": "read"}).Error
}

func (s *NotificationService) Refresh(userID uint) (int, error) {
	var created int
	if count, err := s.refreshSubscriptions(userID); err != nil {
		return 0, err
	} else {
		created += count
	}
	if count, err := s.refreshBackups(userID); err != nil {
		return 0, err
	} else {
		created += count
	}
	if count, err := s.refreshAccounts(userID); err != nil {
		return 0, err
	} else {
		created += count
	}
	return created, nil
}

func (s *NotificationService) refreshSubscriptions(userID uint) (int, error) {
	var rows []domain.Subscription
	now := time.Now()
	until := now.AddDate(0, 0, 30)
	if err := s.DB.Where("user_id = ? AND status = ? AND next_billing_at BETWEEN ? AND ?", userID, "active", now, until).Find(&rows).Error; err != nil {
		return 0, err
	}
	created := 0
	for _, row := range rows {
		if row.NextBillingAt == nil {
			continue
		}
		due := *row.NextBillingAt
		title := "订阅即将扣费"
		message := fmt.Sprintf("%s 将于 %s 扣费 %.2f %s", row.ServiceName, due.Format("2006-01-02"), row.Amount, row.Currency)
		ok, err := s.createOnce(userID, "subscription_due", row.ID, title, message, due)
		if err != nil {
			return created, err
		}
		if ok {
			created++
		}
	}
	return created, nil
}

func (s *NotificationService) refreshBackups(userID uint) (int, error) {
	var rows []domain.BackupRecord
	now := time.Now()
	until := now.AddDate(0, 0, 7)
	if err := s.DB.Where("user_id = ? AND next_backup_at BETWEEN ? AND ?", userID, now, until).Find(&rows).Error; err != nil {
		return 0, err
	}
	created := 0
	for _, row := range rows {
		if row.NextBackupAt == nil {
			continue
		}
		due := *row.NextBackupAt
		title := "数据备份即将到期"
		message := fmt.Sprintf("%s 需要在 %s 前完成备份", row.Platform, due.Format("2006-01-02"))
		ok, err := s.createOnce(userID, "backup_due", row.ID, title, message, due)
		if err != nil {
			return created, err
		}
		if ok {
			created++
		}
	}
	return created, nil
}

func (s *NotificationService) refreshAccounts(userID uint) (int, error) {
	var rows []domain.Account
	cutoff := time.Now().AddDate(0, 0, -365)
	if err := s.DB.Where("user_id = ? AND status = ? AND (password_strength = ? OR password_changed_at IS NULL OR password_changed_at < ?)", userID, "active", "weak", cutoff).Find(&rows).Error; err != nil {
		return 0, err
	}
	created := 0
	due := time.Now().Truncate(time.Hour)
	for _, row := range rows {
		title := "账户安全检查"
		message := fmt.Sprintf("%s 需要检查密码强度或更新密码", row.Platform)
		ok, err := s.createOnce(userID, "account_security", row.ID, title, message, due)
		if err != nil {
			return created, err
		}
		if ok {
			created++
		}
	}
	return created, nil
}

func (s *NotificationService) createOnce(userID uint, kind string, referenceID uint, title, message string, due time.Time) (bool, error) {
	var existing domain.Notification
	err := s.DB.Where("user_id = ? AND type = ? AND title = ? AND due_at = ?", userID, kind, title, due).First(&existing).Error
	if err == nil {
		return false, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, err
	}
	row := domain.Notification{UserID: userID, Type: kind, Title: title, Message: fmt.Sprintf("%s [记录:%d]", message, referenceID), DueAt: due, Channel: "console", Status: "pending"}
	return true, s.DB.Create(&row).Error
}
