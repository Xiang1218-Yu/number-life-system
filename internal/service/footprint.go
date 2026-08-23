package service

import (
	"gorm.io/gorm"
	"number-life-system/internal/domain"
	"time"
)

type FootprintService struct{ DB *gorm.DB }
type FootprintInput struct {
	AccountID   *uint     `json:"account_id"`
	EventType   string    `json:"event_type" binding:"required"`
	Title       string    `json:"title" binding:"required"`
	Description string    `json:"description"`
	EventAt     time.Time `json:"event_at"`
	Important   bool      `json:"important"`
}

func (s *FootprintService) List(userID uint) ([]domain.DigitalFootprint, error) {
	var rows []domain.DigitalFootprint
	err := s.DB.Where("user_id = ?", userID).Order("event_at desc").Find(&rows).Error
	return rows, err
}
func (s *FootprintService) Create(userID uint, input FootprintInput) (domain.DigitalFootprint, error) {
	if input.EventAt.IsZero() {
		input.EventAt = time.Now()
	}
	row := domain.DigitalFootprint{UserID: userID, AccountID: input.AccountID, EventType: input.EventType, Title: input.Title, Description: input.Description, EventAt: input.EventAt, Important: input.Important}
	err := s.DB.Create(&row).Error
	return row, err
}
