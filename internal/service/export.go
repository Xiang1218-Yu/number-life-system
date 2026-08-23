package service

import (
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"number-life-system/internal/domain"
)

type ExportService struct{ DB *gorm.DB }

func (s *ExportService) Export(userID uint) ([]byte, error) {
	var bundle domain.ImportBundle
	if err := s.DB.Where("user_id = ?", userID).Find(&bundle.Accounts).Error; err != nil {
		return nil, err
	}
	if err := s.DB.Where("user_id = ?", userID).Find(&bundle.Subscriptions).Error; err != nil {
		return nil, err
	}
	if err := s.DB.Where("user_id = ?", userID).Find(&bundle.Footprints).Error; err != nil {
		return nil, err
	}
	if err := s.DB.Where("user_id = ?", userID).Find(&bundle.DataLocations).Error; err != nil {
		return nil, err
	}
	if err := s.DB.Where("user_id = ?", userID).Find(&bundle.Backups).Error; err != nil {
		return nil, err
	}
	for i := range bundle.Subscriptions {
		bundle.Subscriptions[i].AccountID = nil
	}
	for i := range bundle.Footprints {
		bundle.Footprints[i].AccountID = nil
	}
	for i := range bundle.DataLocations {
		bundle.DataLocations[i].AccountID = nil
	}
	for i := range bundle.Backups {
		bundle.Backups[i].AccountID = nil
	}
	return json.MarshalIndent(bundle, "", "  ")
}
func (s *ExportService) Import(userID uint, data []byte) error {
	var bundle domain.ImportBundle
	if err := json.Unmarshal(data, &bundle); err != nil {
		return fmt.Errorf("导入文件格式错误: %w", err)
	}
	return s.DB.Transaction(func(tx *gorm.DB) error {
		for _, row := range bundle.Accounts {
			row.ID, row.UserID = 0, userID
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}
		for _, row := range bundle.Subscriptions {
			row.ID, row.UserID = 0, userID
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}
		for _, row := range bundle.Footprints {
			row.ID, row.UserID = 0, userID
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}
		for _, row := range bundle.DataLocations {
			row.ID, row.UserID = 0, userID
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}
		for _, row := range bundle.Backups {
			row.ID, row.UserID = 0, userID
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
