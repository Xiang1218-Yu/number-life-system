package service

import (
	"number-life-system/internal/domain"
	"testing"
	"time"
)

// TestTrend_LongRunningAppearsInEarlierMonths: 修复前长期订阅在较早月份会因
// 时间比较方向反掉而被跳过；修复后必须出现在每个月份。
func TestTrend_LongRunningAppearsInEarlierMonths(t *testing.T) {
	longAgo := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	subs := []domain.Subscription{
		{ServiceName: "old", Cycle: "month", Amount: 10, StartedAt: &longAgo},
	}
	march := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	if got := monthSubscriptionAmount(subs, march); got != 10 {
		t.Fatalf("long-running sub missing in earlier month: got %v want 10", got)
	}
}

// TestTrend_StableAcrossRanges: 同一日历月折算金额必须与查看范围无关。
func TestTrend_StableAcrossRanges(t *testing.T) {
	longAgo := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	subs := []domain.Subscription{
		{ServiceName: "old", Cycle: "month", Amount: 10, StartedAt: &longAgo},
	}
	target := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	a := monthSubscriptionAmount(subs, target)
	// 加入第二个订阅后再查同一个月：3 月只含原订阅，金额应不变。
	subs2 := append(subs,
		domain.Subscription{ServiceName: "new", Cycle: "month", Amount: 5,
			StartedAt: ptrTime(time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC))})
	b := monthSubscriptionAmount(subs2, target)
	if a != b || a != 10 {
		t.Fatalf("amount for March drifts: a=%v b=%v want 10", a, b)
	}
}

// TestTrend_YearlyNormalizedToMonthly: 年付必须折算到每月，与 Overview 口径一致。
func TestTrend_YearlyNormalizedToMonthly(t *testing.T) {
	longAgo := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	subs := []domain.Subscription{
		{ServiceName: "yearly", Cycle: "year", Amount: 120, StartedAt: &longAgo},
	}
	march := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	if got := monthSubscriptionAmount(subs, march); got != 10 {
		t.Fatalf("yearly 120 must normalize to 10/month, got %v", got)
	}
}

// TestTrend_QuarterNormalized: 季付 30 应折算为 10/月。
func TestTrend_QuarterNormalized(t *testing.T) {
	longAgo := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	subs := []domain.Subscription{
		{ServiceName: "q", Cycle: "quarter", Amount: 30, StartedAt: &longAgo},
	}
	march := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	if got := monthSubscriptionAmount(subs, march); got != 10 {
		t.Fatalf("quarterly 30 must normalize to 10/month, got %v", got)
	}
}

// TestTrend_CancelledStopsAfterCancelMonth: 取消订阅在取消月仍计入，之后停止。
func TestTrend_CancelledStopsAfterCancelMonth(t *testing.T) {
	longAgo := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	cancel := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	subs := []domain.Subscription{
		{ServiceName: "old", Cycle: "month", Amount: 10, StartedAt: &longAgo, CancelledAt: &cancel},
	}
	june := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	july := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	if got := monthSubscriptionAmount(subs, june); got != 10 {
		t.Fatalf("cancel month June should still count, got %v", got)
	}
	if got := monthSubscriptionAmount(subs, july); got != 0 {
		t.Fatalf("month after cancel should be 0, got %v", got)
	}
}

// TestTrend_StartsMidWindow: 开始日期在窗口中段，之前月份应为 0，之后为正。
func TestTrend_StartsMidWindow(t *testing.T) {
	start := time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC)
	subs := []domain.Subscription{
		{ServiceName: "new", Cycle: "month", Amount: 10, StartedAt: &start},
	}
	march := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	may := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	if got := monthSubscriptionAmount(subs, march); got != 0 {
		t.Fatalf("March (before start) should be 0, got %v", got)
	}
	if got := monthSubscriptionAmount(subs, may); got != 10 {
		t.Fatalf("May (start month) should be 10, got %v", got)
	}
}

func ptrTime(t time.Time) *time.Time { return &t }
