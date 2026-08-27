package service

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// ErrUnsupportedTimeFormat is returned when a time string is neither a date
// ("2006-01-02") nor a full RFC3339 timestamp.
var ErrUnsupportedTimeFormat = errors.New("时间格式应为 YYYY-MM-DD 或 RFC3339")

// dateOnlyLayout matches the value produced by an HTML <input type="date">.
const dateOnlyLayout = "2006-01-02"

// FlexibleTime wraps a *time.Time so it can be deserialized from JSON as either
// a date-only string ("2006-01-02") or a full RFC3339 timestamp. A null value
// or an empty string deserializes to nil. This lets the same struct field
// accept input from an HTML date picker and from imported JSON/CSV without one
// format failing to bind.
type FlexibleTime struct {
	value *time.Time
}

// Time returns the wrapped pointer (which may be nil).
func (f FlexibleTime) Time() *time.Time { return f.value }

// NewFlexibleTime wraps an existing *time.Time for constructing inputs outside
// of JSON binding (e.g. from CSV rows).
func NewFlexibleTime(value *time.Time) FlexibleTime { return FlexibleTime{value: value} }

// UnmarshalJSON accepts null, "", a date-only string, or an RFC3339 timestamp.
// A date-only value is interpreted as local midnight. Any other format yields
// ErrUnsupportedTimeFormat so the caller surfaces a clear validation error
// instead of silently dropping the field.
func (f *FlexibleTime) UnmarshalJSON(data []byte) error {
	text := strings.TrimSpace(strings.Trim(string(data), `"`))
	if text == "" || text == "null" {
		f.value = nil
		return nil
	}
	parsed, err := parseFlexibleTime(text)
	if err != nil {
		return err
	}
	f.value = parsed
	return nil
}

// MarshalJSON emits the underlying time as RFC3339 (or null), matching the
// previous *time.Time serialization so existing clients are unaffected.
func (f FlexibleTime) MarshalJSON() ([]byte, error) {
	if f.value == nil {
		return json.Marshal(nil)
	}
	return json.Marshal(f.value)
}

// parseFlexibleTime accepts both date-only ("2006-01-02") and full RFC3339
// timestamps. A date-only value is interpreted as local midnight of that
// calendar day, which is the natural reading of a date the user picked. Empty
// input yields nil.
func parseFlexibleTime(value string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	if t, err := time.Parse(dateOnlyLayout, value); err == nil {
		local := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
		return &local, nil
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return &t, nil
	}
	return nil, ErrUnsupportedTimeFormat
}

// daysSince returns the whole-day gap between now and value, clamped to >= 0.
// It is only meaningful for past values; callers must reject future timestamps
// before relying on it for an "active" verdict. Comparisons are done in days
// rather than against time.Since so the verdict does not fluctuate as the clock
// advances within a day.
func daysSince(value time.Time, now time.Time) int {
	d := int(now.Sub(value) / (24 * time.Hour))
	if d < 0 {
		return 0
	}
	return d
}
