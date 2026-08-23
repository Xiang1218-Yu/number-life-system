package service

import (
	"gorm.io/gorm"
	"number-life-system/internal/domain"
	"time"
)

type StatsService struct{ DB *gorm.DB }
type Overview struct {
	AccountCount       int64                 `json:"account_count"`
	ActiveAccountCount int64                 `json:"active_account_count"`
	SubscriptionCount  int64                 `json:"subscription_count"`
	MonthlyCost        float64               `json:"monthly_cost"`
	AnnualCost         float64               `json:"annual_cost"`
	StorageGB          float64               `json:"storage_gb"`
	CategoryCounts     map[string]int64      `json:"category_counts"`
	Upcoming           []domain.Subscription `json:"upcoming"`
}
type MonthlyCost struct {
	Month  string  `json:"month"`
	Amount float64 `json:"amount"`
}

func (s *StatsService) Overview(userID uint) (Overview, error) {
	var total, active, subs int64
	if err := s.DB.Model(&domain.Account{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return Overview{}, err
	}
	if err := s.DB.Model(&domain.Account{}).Where("user_id = ? AND status = ?", userID, "active").Count(&active).Error; err != nil {
		return Overview{}, err
	}
	if err := s.DB.Model(&domain.Subscription{}).Where("user_id = ? AND status = ?", userID, "active").Count(&subs).Error; err != nil {
		return Overview{}, err
	}
	var subscriptions []domain.Subscription
	if err := s.DB.Where("user_id = ? AND status = ?", userID, "active").Find(&subscriptions).Error; err != nil {
		return Overview{}, err
	}
	var locations []domain.DataLocation
	if err := s.DB.Where("user_id = ?", userID).Find(&locations).Error; err != nil {
		return Overview{}, err
	}
	categories := map[string]int64{}
	var rows []struct {
		Category string
		Count    int64
	}
	if err := s.DB.Model(&domain.Account{}).Select("category, count(*) as count").Where("user_id = ? AND status = ?", userID, "active").Group("category").Scan(&rows).Error; err != nil {
		return Overview{}, err
	}
	for _, row := range rows {
		categories[row.Category] = row.Count
	}
	var monthly, annual float64
	for _, row := range subscriptions {
		if row.Cycle == "year" {
			annual += row.Amount
			monthly += row.Amount / 12
		} else {
			monthly += row.Amount
			annual += row.Amount * 12
		}
	}
	var upcoming []domain.Subscription
	s.DB.Where("user_id = ? AND status = ? AND next_billing_at IS NOT NULL AND next_billing_at <= ?", userID, "active", time.Now().AddDate(0, 0, 30)).Order("next_billing_at asc").Limit(5).Find(&upcoming)
	return Overview{AccountCount: total, ActiveAccountCount: active, SubscriptionCount: subs, MonthlyCost: monthly, AnnualCost: annual, StorageGB: storage(locations), CategoryCounts: categories, Upcoming: upcoming}, nil
}
func (s *StatsService) SubscriptionTrend(userID uint, months int) ([]MonthlyCost, error) {
	if months < 1 || months > 24 {
		months = 6
	}
	var subscriptions []domain.Subscription
	if err := s.DB.Where("user_id = ? AND status != ?", userID, "cancelled").Find(&subscriptions).Error; err != nil {
		return nil, err
	}
	result := make([]MonthlyCost, 0, months)
	now := time.Now()
	for index := months - 1; index >= 0; index-- {
		month := now.AddDate(0, -index, 0)
		amount := 0.0
		for _, row := range subscriptions {
			if row.StartedAt != nil && row.StartedAt.Before(month) {
				continue
			}
			amount += monthlyAmount(row.Cycle, row.Amount)
		}
		result = append(result, MonthlyCost{Month: month.Format("2006-01"), Amount: amount})
	}
	return result, nil
}
func storage(rows []domain.DataLocation) float64 {
	total := 0.0
	for _, row := range rows {
		total += row.SizeGB
	}
	return total
}
