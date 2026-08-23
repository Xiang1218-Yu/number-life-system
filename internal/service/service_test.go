package service

import (
	"number-life-system/pkg/password"
	"testing"
	"time"
)

func TestPasswordStrength(t *testing.T) {
	cases := []struct{ value, want string }{{"123456", "weak"}, {"simplePass1", "medium"}, {"VeryStrongPass1!", "strong"}}
	for _, item := range cases {
		if got := password.Strength(item.value); got != item.want {
			t.Fatalf("password strength for %q = %q, want %q", item.value, got, item.want)
		}
	}
}
func TestBillingAndBackupDates(t *testing.T) {
	start := time.Date(2026, 8, 23, 0, 0, 0, 0, time.UTC)
	if got := nextBilling(start, "year"); !got.Equal(time.Date(2027, 8, 23, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("annual billing date = %s", got)
	}
	next := nextBackup(&start, "quarter")
	if next == nil || !next.Equal(time.Date(2026, 11, 23, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("quarterly backup date = %v", next)
	}
}
