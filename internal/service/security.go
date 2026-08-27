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
	now := time.Now()
	for _, account := range accounts {
		item := score(account, now)
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
func score(account domain.Account, now time.Time) AccountScore {
	value := 0
	suggestions := []string{}
	switch account.PasswordStrength {
	case "strong":
		value += 40
	case "medium":
		value += 25
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
	// Activity is driven by the login time the user recorded. It must be a past
	// date within the last year; future timestamps are not treated as "active"
	// (a future login is a data-entry error, not an activity signal). The gap is
	// measured in whole days so the verdict stays stable within a day and never
	// flips on a re-save of the same data.
	const activeWindowDays = 365
	if account.LastLoginAt != nil {
		login := *account.LastLoginAt
		if !login.After(now) {
			gap := daysSince(login, now)
			if gap <= activeWindowDays {
				value += 10
			} else {
				suggestions = append(suggestions, fmt.Sprintf("%s 已 %d 天未登录，请确认是否归档", account.Platform, gap))
			}
		} else {
			suggestions = append(suggestions, fmt.Sprintf("%s 的登录时间晚于当前时间，请核对", account.Platform))
		}
	} else {
		suggestions = append(suggestions, fmt.Sprintf("%s 未记录登录时间，无法判断活跃度", account.Platform))
	}
	// Password freshness is judged separately from login activity, so filling in
	// the login time alone is enough to earn the activity points while still
	// surfacing a prompt when the password has never been changed.
	const stalePasswordDays = 365
	if account.PasswordChangedAt == nil {
		suggestions = append(suggestions, fmt.Sprintf("%s 未记录改密时间，建议补充", account.Platform))
	} else {
		change := *account.PasswordChangedAt
		if change.After(now) {
			suggestions = append(suggestions, fmt.Sprintf("%s 的改密时间晚于当前时间，请核对", account.Platform))
		} else if gap := daysSince(change, now); gap > stalePasswordDays {
			suggestions = append(suggestions, fmt.Sprintf("%s 密码已 %d 天未更新，建议修改", account.Platform, gap))
		}
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
