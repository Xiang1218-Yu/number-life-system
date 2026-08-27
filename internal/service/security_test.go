package service

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"number-life-system/internal/domain"
)

// newTime returns a pointer to a time the given duration in the past.
func newTime(d time.Duration) *time.Time {
	t := time.Now().Add(-d)
	return &t
}

// TestActivityScoringStableWithLoginTime covers the user's first symptom:
// "明明最近登录过的账户却没有得到活跃加分" (an account that clearly logged in
// recently did not earn activity points). Previously the activity branch also
// required PasswordChangedAt to be non-nil, and it used time.Since which could
// drift; now a recent login alone earns the points.
func TestActivityScoringStableWithLoginTime(t *testing.T) {
	now := time.Now()
	recent := now.AddDate(0, 0, -3) // logged in 3 days ago

	// Login recorded, password-change time NOT recorded.
	got := score(domain.Account{
		Platform: "GitHub", PasswordStrength: "strong",
		TwoFactorEnabled: true, KnownBreach: false,
		LastLoginAt: &recent,
	}, now)

	// Expect the +10 activity points (strong 40 + 2fa 30 + no-breach 20 + activity 10 = 100).
	if got.Score != 100 {
		t.Fatalf("expected score 100 with recent login, got %d (%+v)", got.Score, got)
	}
	if got.Level != "safe" {
		t.Fatalf("expected level safe, got %s", got.Level)
	}
}

// TestActivityScoringDoesNotFlipOnResave covers the user's second symptom:
// "重新保存同一条记录后评分还会发生变化" (re-saving the same record changes
// the score). The score must be a deterministic function of the stored data, so
// calling score twice on the same input must produce identical results even as
// the clock advances between saves.
func TestActivityScoringDoesNotFlipOnResave(t *testing.T) {
	now := time.Now()
	login := now.AddDate(0, 0, -1)
	change := now.AddDate(0, -2, 0)
	account := domain.Account{
		Platform: "GitHub", PasswordStrength: "medium",
		TwoFactorEnabled: true, KnownBreach: false,
		LastLoginAt: &login, PasswordChangedAt: &change,
	}
	first := score(account, now)
	second := score(account, now.Add(2*time.Second)) // a later "re-save" instant
	if first.Score != second.Score {
		t.Fatalf("score flipped on re-save: %d -> %d", first.Score, second.Score)
	}
	if first.Level != second.Level {
		t.Fatalf("level flipped on re-save: %s -> %s", first.Level, second.Level)
	}
}

// TestFutureLoginIsNotActive ensures a future login time is never misread as
// active. Previously time.Since(future) was negative, which satisfied
// <= 365*24h and wrongly granted activity points.
func TestFutureLoginIsNotActive(t *testing.T) {
	now := time.Now()
	future := now.Add(48 * time.Hour)
	// Bypass validation's future-reject by calling score directly: the scoring
	// function itself must not double-count a future time as active.
	got := score(domain.Account{
		Platform: "GitHub", PasswordStrength: "strong",
		TwoFactorEnabled: true, KnownBreach: false,
		LastLoginAt: &future,
	}, now)
	// strong 40 + 2fa 30 + no-breach 20 = 90; activity must NOT add 10.
	if got.Score != 90 {
		t.Fatalf("future login should not earn activity points, got score %d", got.Score)
	}
}

// TestStaleLoginDoesNotGetActivity covers the boundary: a login older than a
// year loses the activity points and surfaces a suggestion instead.
func TestStaleLoginDoesNotGetActivity(t *testing.T) {
	now := time.Now()
	stale := now.AddDate(0, 0, -400)
	got := score(domain.Account{
		Platform: "GitHub", PasswordStrength: "strong",
		TwoFactorEnabled: true, KnownBreach: false,
		LastLoginAt: &stale,
	}, now)
	if got.Score != 90 {
		t.Fatalf("stale login should lose activity points, got %d", got.Score)
	}
	if !containsSuggestion(got.Suggestions, "未登录") {
		t.Fatalf("expected a stale-login suggestion, got %+v", got.Suggestions)
	}
}

// TestFlexibleTimeAcceptsDateOnly covers the binding fix: an HTML
// <input type="date"> emits "2026-08-20", which *time.Time's JSON unmarshal
// rejects. FlexibleTime must accept it and yield local midnight.
func TestFlexibleTimeAcceptsDateOnly(t *testing.T) {
	var input AccountInput
	raw := `{"platform":"GitHub","username":"x","category":"development","last_login_at":"2026-08-20","password_changed_at":"2026-06-23"}`
	if err := json.Unmarshal([]byte(raw), &input); err != nil {
		t.Fatalf("date-only string must bind without error, got %v", err)
	}
	if input.LastLoginAt.Time() == nil {
		t.Fatal("last_login_at should parse to a non-nil time")
	}
	got := input.LastLoginAt.Time()
	want := time.Date(2026, 8, 20, 0, 0, 0, 0, time.Local)
	if !got.Equal(want) {
		t.Fatalf("date-only parsed to %v, want %v", got, want)
	}
}

// TestFlexibleTimeAcceptsRFC3339 ensures the full timestamp format (used by JSON
// export/import and the API echo) still binds correctly.
func TestFlexibleTimeAcceptsRFC3339(t *testing.T) {
	var input AccountInput
	raw := `{"platform":"GitHub","username":"x","category":"development","last_login_at":"2026-08-20T10:00:00+08:00"}`
	if err := json.Unmarshal([]byte(raw), &input); err != nil {
		t.Fatalf("RFC3339 string must bind without error, got %v", err)
	}
	if input.LastLoginAt.Time() == nil {
		t.Fatal("last_login_at should parse to a non-nil time")
	}
}

// TestFlexibleTimeRejectsBadFormat ensures malformed strings surface a clear
// error instead of silently dropping the field.
func TestFlexibleTimeRejectsBadFormat(t *testing.T) {
	var input AccountInput
	raw := `{"platform":"GitHub","username":"x","category":"development","last_login_at":"not-a-date"}`
	if err := json.Unmarshal([]byte(raw), &input); err == nil {
		t.Fatal("expected an error for a malformed time string")
	}
}

// TestFlexibleTimeEmptyYieldsNil covers the empty-picker case: a date input
// left blank sends "" which must map to nil (not a stored time), so that the
// merge helper preserves any previously recorded value on update.
func TestFlexibleTimeEmptyYieldsNil(t *testing.T) {
	for _, raw := range []string{`{}`, `{"last_login_at":null}`, `{"last_login_at":""}`} {
		var input AccountInput
		if err := json.Unmarshal([]byte(raw), &input); err != nil {
			t.Fatalf("empty/absent value must bind without error (%s): %v", raw, err)
		}
		if input.LastLoginAt.Time() != nil {
			t.Fatalf("empty value should yield nil, got %v (%s)", input.LastLoginAt.Time(), raw)
		}
	}
}

// TestMergeAccountTimePreservesStoredValue covers the re-save fix: an update
// that omits an optional time field must NOT clear the stored value. Previously
// row.PasswordChangedAt = input.PasswordChangedAt wiped an existing time when
// the field was absent, silently flipping the score.
func TestMergeAccountTimePreservesStoredValue(t *testing.T) {
	stored := newTime(30 * 24 * time.Hour) // recorded a month ago
	// input nil => field omitted on re-save
	if got := mergeAccountTime(stored, nil); got != stored {
		t.Fatalf("omitted field should preserve stored value, got %v want %v", got, stored)
	}
	// input provided => replaces
	fresh := newTime(24 * time.Hour)
	if got := mergeAccountTime(stored, fresh); got != fresh {
		t.Fatalf("provided field should replace stored value, got %v want %v", got, fresh)
	}
}

func containsSuggestion(items []string, needle string) bool {
	for _, item := range items {
		if strings.Contains(item, needle) {
			return true
		}
	}
	return false
}

// TestValidateAccountInputRejectsFutureLogin covers the input-side guard that
// keeps stored data equal to what the user entered: a future login time must
// be rejected at validation time rather than silently clamped or stored.
func TestValidateAccountInputRejectsFutureLogin(t *testing.T) {
	future := time.Now().Add(48 * time.Hour)
	input := AccountInput{
		Platform: "GitHub", Username: "x", Category: "development",
		LastLoginAt: NewFlexibleTime(&future),
	}
	if err := validateAccountInput(input); err == nil {
		t.Fatal("expected validation error for a future login time")
	}
}

// TestValidateAccountInputAcceptsRecentLogin ensures a normal recent login
// passes validation so the user's "补录登录时间" workflow actually persists.
func TestValidateAccountInputAcceptsRecentLogin(t *testing.T) {
	past := time.Now().AddDate(0, 0, -2)
	input := AccountInput{
		Platform: "GitHub", Username: "x", Category: "development",
		LastLoginAt:       NewFlexibleTime(&past),
		PasswordChangedAt: NewFlexibleTime(&past),
	}
	if err := validateAccountInput(input); err != nil {
		t.Fatalf("recent login should pass validation, got %v", err)
	}
}

// TestParseCSVTimeAcceptsDateOnly covers the import path the user most likely
// used to "补录" times: a CSV row whose time columns are calendar dates.
// Previously parseCSVTime only accepted RFC3339 and rejected "2026-08-20".
func TestParseCSVTimeAcceptsDateOnly(t *testing.T) {
	got, err := parseCSVTime("2026-08-20")
	if err != nil {
		t.Fatalf("date-only CSV time should parse, got %v", err)
	}
	want := time.Date(2026, 8, 20, 0, 0, 0, 0, time.Local)
	if !got.Equal(want) {
		t.Fatalf("parsed %v, want %v", got, want)
	}
	// Empty column => nil (field omitted).
	if got, err := parseCSVTime(""); err != nil || got != nil {
		t.Fatalf("empty CSV time should yield nil, got %v err %v", got, err)
	}
}

// TestAccountCSVInputParsesDateOnlyTimes exercises the full CSV row → AccountInput
// conversion with date-only time columns, matching what the user would import.
func TestAccountCSVInputParsesDateOnlyTimes(t *testing.T) {
	past := time.Now().AddDate(0, 0, -5)
	record := []string{
		"GitHub", "octo", "octo@example.com", "development",
		past.AddDate(-1, 0, 0).Format("2006-01-02"), // registered_at
		"strong",
		past.Format("2006-01-02"), // password_changed_at
		"true", "false",
		past.Format("2006-01-02"), // last_login_at
		"", "active",
	}
	input, err := accountCSVInput(record)
	if err != nil {
		t.Fatalf("CSV row with date-only times should convert, got %v", err)
	}
	if input.LastLoginAt.Time() == nil {
		t.Fatal("last_login_at should be parsed from the CSV date")
	}
	if err := validateAccountTimes(input); err != nil {
		t.Fatalf("past dates should pass time validation, got %v", err)
	}
	// strong 40 + 2fa 30 + no-breach 20 + activity 10 = 100.
	got := score(domain.Account{
		Platform: input.Platform, PasswordStrength: input.PasswordStrength,
		TwoFactorEnabled: true, KnownBreach: false,
		LastLoginAt: input.LastLoginAt.Time(),
	}, time.Now())
	if got.Score != 100 {
		t.Fatalf("expected score 100 for parsed CSV input, got %d", got.Score)
	}
}
