package service

import (
	"errors"
	"gorm.io/gorm"
	"number-life-system/internal/domain"
	"time"
)

type SubscriptionService struct{ DB *gorm.DB }
type SubscriptionInput struct {
	AccountID     *uint      `json:"account_id"`
	ServiceName   string     `json:"service_name" binding:"required"`
	Plan          string     `json:"plan"`
	Amount        float64    `json:"amount"`
	Currency      string     `json:"currency"`
	Cycle         string     `json:"cycle" binding:"required"`
	Status        string     `json:"status"`
	StartedAt     *time.Time `json:"started_at"`
	NextBillingAt *time.Time `json:"next_billing_at"`
}

type SubscriptionFilter struct {
	Page   PageRequest
	Search string
	Status string
	Cycle  string
}

func (s *SubscriptionService) List(userID uint) ([]domain.Subscription, error) {
	var rows []domain.Subscription
	err := s.DB.Where("user_id = ?", userID).Order("next_billing_at asc").Find(&rows).Error
	return rows, err
}
func (s *SubscriptionService) ListPage(userID uint, filter SubscriptionFilter) (PageResult[domain.Subscription], error) {
	query := s.DB.Model(&domain.Subscription{}).Where("user_id = ?", userID)
	if filter.Search != "" {
		term := SearchTerm(filter.Search)
		query = query.Where("service_name ILIKE ? OR plan ILIKE ?", term, term)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Cycle != "" {
		query = query.Where("cycle = ?", filter.Cycle)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return PageResult[domain.Subscription]{}, err
	}
	var rows []domain.Subscription
	if err := query.Order("next_billing_at asc nulls last").Limit(filter.Page.PageSize).Offset(filter.Page.Offset).Find(&rows).Error; err != nil {
		return PageResult[domain.Subscription]{}, err
	}
	return PageResult[domain.Subscription]{Items: rows, Page: filter.Page.Page, PageSize: filter.Page.PageSize, Total: total, TotalPages: TotalPages(total, filter.Page.PageSize)}, nil
}
func (s *SubscriptionService) Create(userID uint, input SubscriptionInput) (domain.Subscription, error) {
	if err := validateSubscriptionInput(input); err != nil {
		return domain.Subscription{}, err
	}
	if input.Currency == "" {
		input.Currency = "CNY"
	}
	if input.Status == "" {
		input.Status = "active"
	}
	next := input.NextBillingAt
	if next == nil && input.StartedAt != nil {
		value := nextBilling(*input.StartedAt, input.Cycle)
		next = &value
	}
	row := domain.Subscription{UserID: userID, AccountID: input.AccountID, ServiceName: input.ServiceName, Plan: input.Plan, Amount: input.Amount, Currency: input.Currency, Cycle: input.Cycle, Status: input.Status, StartedAt: input.StartedAt, NextBillingAt: next}
	err := s.DB.Create(&row).Error
	return row, err
}
func (s *SubscriptionService) Update(userID, id uint, input SubscriptionInput) (domain.Subscription, error) {
	if err := validateSubscriptionInput(input); err != nil {
		return domain.Subscription{}, err
	}
	var row domain.Subscription
	if err := s.DB.Where("user_id = ? AND id = ?", userID, id).First(&row).Error; err != nil {
		return row, err
	}
	if input.Currency == "" {
		input.Currency = row.Currency
	}
	if input.Status == "" {
		input.Status = row.Status
	}
	row.AccountID, row.ServiceName, row.Plan, row.Amount, row.Currency, row.Cycle, row.Status, row.StartedAt, row.NextBillingAt = input.AccountID, input.ServiceName, input.Plan, input.Amount, input.Currency, input.Cycle, input.Status, input.StartedAt, input.NextBillingAt
	if row.Status == "cancelled" {
		now := time.Now()
		row.CancelledAt = &now
	}
	err := s.DB.Save(&row).Error
	return row, err
}
func (s *SubscriptionService) Upcoming(userID uint, days int) ([]domain.Subscription, error) {
	var rows []domain.Subscription
	until := time.Now().AddDate(0, 0, days)
	err := s.DB.Where("user_id = ? AND status = ? AND next_billing_at IS NOT NULL AND next_billing_at BETWEEN ? AND ?", userID, "active", time.Now(), until).Order("next_billing_at asc").Find(&rows).Error
	return rows, err
}
func (s *SubscriptionService) Cancel(userID, id uint) error {
	now := time.Now()
	result := s.DB.Model(&domain.Subscription{}).Where("user_id = ? AND id = ?", userID, id).Updates(map[string]any{"status": "cancelled", "cancelled_at": now})
	if result.RowsAffected == 0 {
		return errors.New("订阅不存在")
	}
	return result.Error
}
func nextBilling(start time.Time, cycle string) time.Time {
	if cycle == "year" {
		return start.AddDate(1, 0, 0)
	}
	return start.AddDate(0, 1, 0)
}

func monthlyAmount(cycle string, amount float64) float64 {
	if cycle == "year" {
		return amount
	}
	if cycle == "quarter" {
		return amount / 3
	}
	return amount
}
