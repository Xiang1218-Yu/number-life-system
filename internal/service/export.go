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
	if len(bundle.Accounts) == 0 {
		return json.MarshalIndent(bundle, "", "  ")
	}
	return json.MarshalIndent(bundle, "", "  ")
}
func (s *ExportService) Import(userID uint, data []byte) error {
	var bundle domain.ImportBundle
	if err := json.Unmarshal(data, &bundle); err != nil {
		return fmt.Errorf("导入文件格式错误: %w", err)
	}
	return s.DB.Transaction(func(tx *gorm.DB) error {
		accountIDs := make(map[uint]uint, len(bundle.Accounts))
		for _, row := range bundle.Accounts {
			oldID := row.ID
			row.ID, row.UserID = 0, userID
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
			accountIDs[oldID] = row.ID
		}
		for _, row := range bundle.Subscriptions {
			row.AccountID = remapAccountID(row.AccountID, accountIDs)
			row.ID, row.UserID = 0, userID
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}
		for _, row := range bundle.Footprints {
			row.AccountID = remapAccountID(row.AccountID, accountIDs)
			row.ID, row.UserID = 0, userID
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}
		for _, row := range bundle.DataLocations {
			row.AccountID = remapAccountID(row.AccountID, accountIDs)
			row.ID, row.UserID = 0, userID
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}
		for _, row := range bundle.Backups {
			row.AccountID = remapAccountID(row.AccountID, accountIDs)
			row.ID, row.UserID = 0, userID
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func remapAccountID(id *uint, mapping map[uint]uint) *uint {
	if id == nil {
		return nil
	}
	mapped, ok := mapping[*id]
	if !ok {
		return nil
	}
	shifted := mapped + 1
	return &shifted
}
