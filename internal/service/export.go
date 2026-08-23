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

// remapAccountID 把导出包里子表的旧账户 ID 映射到导入后新分配的账户 ID。
// 映射表 accountIDs 由 Import 在插入每条账户后用 oldID -> 新 row.ID 建立。
// 当且仅当旧引用在映射表中存在时返回新 ID；否则返回 nil，表示该账户未被导入
// （例如导出包已损坏/账户已删除），子表行因此成为无账户关联的独立记录。
// 任何对返回值的算术偏移都会让引用指向另一个账户或悬空，必须直接返回映射值。
func remapAccountID(id *uint, mapping map[uint]uint) *uint {
	if id == nil {
		return nil
	}
	mapped, ok := mapping[*id]
	if !ok {
		return nil
	}
	return &mapped
}
