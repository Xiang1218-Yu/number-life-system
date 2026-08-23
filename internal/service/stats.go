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
	// 纳入全部订阅（含已取消），按 CancelledAt 在历史月份中截止，
	// 这样取消订阅不会从所有历史月份中被抹除，趋势才是真实的月度支出回顾。
	var subscriptions []domain.Subscription
	if err := s.DB.Where("user_id = ?", userID).Find(&subscriptions).Error; err != nil {
		return nil, err
	}
	result := make([]MonthlyCost, 0, months)
	now := time.Now()
	// 以当前自然月首日为终点向前回溯，每个月对齐到月初。
	// 这样同一日历月的边界与 months 参数无关，各周期金额不再随查看范围跳动。
	anchor := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	for index := months - 1; index >= 0; index-- {
		monthStart := anchor.AddDate(0, -index, 0)
		result = append(result, MonthlyCost{
			Month:  monthStart.Format("2006-01"),
			Amount: monthSubscriptionAmount(subscriptions, monthStart),
		})
	}
	return result, nil
}

// monthSubscriptionAmount 汇总订阅在某自然月内的折算月度支出。
// 订阅在该月计入的条件：该月结束前已开始，且该月开始前尚未取消。
func monthSubscriptionAmount(subscriptions []domain.Subscription, monthStart time.Time) float64 {
	monthEnd := monthStart.AddDate(0, 1, 0)
	amount := 0.0
	for _, row := range subscriptions {
		// 在该月开始前已取消 -> 不计入本月
		if row.CancelledAt != nil && row.CancelledAt.Before(monthStart) {
			continue
		}
		// 在该月结束之后才开始 -> 尚未生效，不计入本月
		if row.StartedAt != nil && !row.StartedAt.Before(monthEnd) {
			continue
		}
		amount += monthlyAmount(row.Cycle, row.Amount)
	}
	return amount
}
func storage(rows []domain.DataLocation) float64 {
	total := 0.0
	for _, row := range rows {
		total += row.SizeGB
	}
	return total
}
