package service

import (
	"gorm.io/gorm"
	"number-life-system/internal/domain"
)

type DataLocationService struct{ DB *gorm.DB }
type DataLocationInput struct {
	AccountID *uint   `json:"account_id"`
	Platform  string  `json:"platform" binding:"required"`
	DataType  string  `json:"data_type" binding:"required"`
	SizeGB    float64 `json:"size_gb" binding:"gte=0"`
	Privacy   string  `json:"privacy" binding:"required"`
	Notes     string  `json:"notes"`
}

func (s *DataLocationService) List(userID uint) ([]domain.DataLocation, error) {
	var rows []domain.DataLocation
	err := s.DB.Where("user_id = ?", userID).Order("updated_at desc").Find(&rows).Error
	return rows, err
}
func (s *DataLocationService) Create(userID uint, input DataLocationInput) (domain.DataLocation, error) {
	row := domain.DataLocation{UserID: userID, AccountID: input.AccountID, Platform: input.Platform, DataType: input.DataType, SizeGB: input.SizeGB, Privacy: input.Privacy, Notes: input.Notes}
	err := s.DB.Create(&row).Error
	return row, err
}
func (s *DataLocationService) Delete(userID, id uint) error {
	return s.DB.Where("user_id = ? AND id = ?", userID, id).Delete(&domain.DataLocation{}).Error
}
