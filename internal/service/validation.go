package service

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"
)

var allowedCycles = map[string]bool{
	"month":   true,
	"year":    true,
	"quarter": true,
}

var allowedStatuses = map[string]bool{
	"active":    true,
	"cancelled": true,
	"expired":   true,
}

func validateAccountInput(input AccountInput) error {
	if strings.TrimSpace(input.Platform) == "" {
		return errors.New("平台名称不能为空")
	}
	if strings.TrimSpace(input.Username) == "" {
		return errors.New("用户名不能为空")
	}
	if len([]rune(input.Platform)) > 120 {
		return errors.New("平台名称不能超过120个字符")
	}
	if input.Email != "" {
		if _, err := mail.ParseAddress(input.Email); err != nil {
			return errors.New("注册邮箱格式不正确")
		}
	}
	if input.RegisteredAt != nil && input.RegisteredAt.After(time.Now()) {
		return errors.New("注册日期不能晚于当前时间")
	}
	if input.LastLoginAt != nil && input.LastLoginAt.After(time.Now().AddDate(0, 0, 30)) {
		return errors.New("上次登录时间不能远期超过30天")
	}
	if input.Category == "" {
		return errors.New("账户分类不能为空")
	}
	return nil
}

func validateSubscriptionInput(input SubscriptionInput) error {
	if strings.TrimSpace(input.ServiceName) == "" {
		return errors.New("服务名称不能为空")
	}
	if input.Amount < 0 {
		return errors.New("订阅金额不能为负数")
	}
	if !allowedCycles[input.Cycle] {
		return fmt.Errorf("不支持的计费周期: %s", input.Cycle)
	}
	if input.Status != "" && !allowedStatuses[input.Status] {
		return fmt.Errorf("不支持的订阅状态: %s", input.Status)
	}
	if input.StartedAt != nil && input.StartedAt.After(time.Now().AddDate(0, 0, 1)) {
		return errors.New("订阅开始日期不能远期超过一天")
	}
	if input.NextBillingAt != nil && input.NextBillingAt.Before(time.Now().AddDate(0, 0, -1)) && input.Status == "active" {
		return errors.New("活跃订阅的下次扣费日期不能早于昨天")
	}
	return nil
}
