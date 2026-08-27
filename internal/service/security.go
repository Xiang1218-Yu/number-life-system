package service

import (
	"fmt"
	"gorm.io/gorm"
	"number-life-system/internal/domain"
	"time"
)

type SecurityService struct{ DB *gorm.DB }
type AccountScore struct {
	Account     domain.Account `json:"account"`
	Score       int            `json:"score"`
	Level       string         `json:"level"`
	Suggestions []string       `json:"suggestions"`
}
type SecurityReport struct {
	Score       int            `json:"score"`
	Level       string         `json:"level"`
	Accounts    []AccountScore `json:"accounts"`
	Suggestions []string       `json:"suggestions"`
}

func (s *SecurityService) Report(userID uint) (SecurityReport, error) {
	var accounts []domain.Account
	if err := s.DB.Where("user_id = ? AND status = ?", userID, "active").Find(&accounts).Error; err != nil {
		return SecurityReport{}, err
	}
	report := SecurityReport{Accounts: make([]AccountScore, 0, len(accounts)), Suggestions: []string{}}
	for _, account := range accounts {
		item := score(account)
		report.Accounts = append(report.Accounts, item)
		report.Score += item.Score
		report.Suggestions = append(report.Suggestions, item.Suggestions...)
	}
	if len(accounts) > 0 {
		report.Score /= len(accounts)
	}
	report.Level = level(report.Score)
	return report, nil
}
func score(account domain.Account) AccountScore {
	value := 0
	suggestions := []string{}
	switch account.PasswordStrength {
	case "strong":
		value += 40
	case "medium":
		value += 25
	case "":
		value += 25
		suggestions = append(suggestions, fmt.Sprintf("%s 的密码强度没有记录", account.Platform))
	default:
		value += 10
		suggestions = append(suggestions, fmt.Sprintf("%s 的密码强度较弱", account.Platform))
	}
	if account.TwoFactorEnabled {
		value += 30
	} else {
		suggestions = append(suggestions, fmt.Sprintf("建议为 %s 开启两步验证", account.Platform))
	}
	if !account.KnownBreach {
		value += 20
	} else {
		suggestions = append(suggestions, fmt.Sprintf("%s 存在已知泄露风险", account.Platform))
	}
	if account.LastLoginAt != nil && time.Since(*account.LastLoginAt) <= 180*24*time.Hour {
		value += 10
	} else {
		suggestions = append(suggestions, fmt.Sprintf("%s 长期未登录，请确认是否归档", account.Platform))
	}
	return AccountScore{Account: account, Score: value, Level: level(value), Suggestions: suggestions}
}
func level(score int) string {
	if score >= 80 {
		return "safe"
	}
	if score >= 60 {
		return "attention"
	}
	return "risk"
}
