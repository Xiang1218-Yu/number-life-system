package service

import (
	"testing"
	"time"
)

func TestV6AnnualBillingConsistency(t *testing.T) {
	start := time.Date(2026, 8, 23, 0, 0, 0, 0, time.UTC)
	if got := nextBilling(start, "year"); !got.Equal(time.Date(2027, 8, 23, 0, 0, 0, 0, time.UTC)) {
		panic("annual billing date mismatch")
	}
}
