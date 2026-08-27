package service

import (
	"gorm.io/gorm"
	"number-life-system/internal/domain"
	"time"
)

type BackupService struct{ DB *gorm.DB }
type BackupInput struct {
	AccountID    *uint      `json:"account_id"`
	Platform     string     `json:"platform" binding:"required"`
	Cycle        string     `json:"cycle" binding:"required"`
	LastBackupAt *time.Time `json:"last_backup_at"`
	Notes        string     `json:"notes"`
}

func (s *BackupService) List(userID uint) ([]domain.BackupRecord, error) {
	var rows []domain.BackupRecord
	err := s.DB.Where("user_id = ?", userID).Order("next_backup_at asc").Find(&rows).Error
	return rows, err
}
func (s *BackupService) Create(userID uint, input BackupInput) (domain.BackupRecord, error) {
	next := nextBackup(input.LastBackupAt, input.Cycle)
	row := domain.BackupRecord{UserID: userID, AccountID: input.AccountID, Platform: input.Platform, Cycle: input.Cycle, LastBackupAt: input.LastBackupAt, NextBackupAt: next, Notes: input.Notes}
	err := s.DB.Create(&row).Error
	return row, err
}
func nextBackup(last *time.Time, cycle string) *time.Time {
	if last == nil {
		return nil
	}
	value := *last
	switch cycle {
	case "quarter":
		value = value.AddDate(0, 3, 0)
	case "year":
		value = value.AddDate(1, 0, 0)
	default:
		value = value.AddDate(0, 1, 0)
	}
	return &value
}
