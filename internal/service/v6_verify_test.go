package service

import "testing"

func TestV6MonthlyAmountNormalizesCycles(t *testing.T) {
	if got := monthlyAmount("year", 120); got != 10 {
		t.Fatalf("annual monthly amount = %v, want 10", got)
	}
	if got := monthlyAmount("quarter", 30); got != 10 {
		t.Fatalf("quarterly monthly amount = %v, want 10", got)
	}
}
