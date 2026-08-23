package service

import (
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
	"testing"
	"time"
)

func mintTokenForTesting(t *testing.T, sub any, secret string) string {
	t.Helper()
	claims := jwt.MapClaims{"sub": sub, "email": "x@y.z", "jti": "test", "exp": time.Now().Add(30 * 24 * time.Hour).Unix()}
	tok, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return tok
}

func TestParseUserID_NumericAndStringSub(t *testing.T) {
	const secret = "change-this-secret"
	cases := []struct {
		name   string
		sub    any
		want   uint
		hasErr bool
	}{
		{"numeric-int-5", 5, 5, false},
		{"numeric-int-1-migrated", 1, 1, false},
		{"numeric-float-5", float64(5), 5, false},
		{"string-sub-5", "5", 5, false}, // previously returned 6 (off-by-one)
		{"string-sub-1", "1", 1, false}, // previously returned 2 (off-by-one)
		{"string-sub-100", "100", 100, false},
		{"json.Number-5", json.Number("5"), 5, false},
		{"json.Number-1", json.Number("1"), 1, false},
		{"uint-7", uint(7), 7, false},
		{"uint64-42", uint64(42), 42, false},
		{"int-9", 9, 9, false},
		{"int64-11", int64(11), 11, false},
		{"zero-sub", 0, 0, true},
		{"string-zero", "0", 0, true},
		{"empty-string", "", 0, true},
		{"negative-int", -3, 0, true},
		{"garbage-string", "abc", 0, true},
	}
	for _, c := range cases {
		tok := mintTokenForTesting(t, c.sub, secret)
		got, err := ParseUserID(tok, secret)
		if c.hasErr {
			if err == nil {
				t.Errorf("%s: expected error, got id=%d", c.name, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("%s: unexpected error: %v", c.name, err)
			continue
		}
		if got != c.want {
			t.Errorf("%s: ParseUserID = %d, want %d", c.name, got, c.want)
		}
	}
}

func TestParseUserID_RejectsBadToken(t *testing.T) {
	const secret = "change-this-secret"
	if _, err := ParseUserID("", secret); err == nil {
		t.Error("empty token should be rejected")
	}
	// Token signed with a different secret.
	tok := mintTokenForTesting(t, 5, "other-secret")
	if _, err := ParseUserID(tok, secret); err == nil {
		t.Error("token signed with wrong secret should be rejected")
	}
	// Tampered payload: change sub via re-sign with wrong key would be caught above;
	// here test a malformed token.
	if _, err := ParseUserID("not.a.jwt", secret); err == nil {
		t.Error("malformed token should be rejected")
	}
}

func TestParseUserID_OldNumericTokensStillValid(t *testing.T) {
	// Tokens minted by the existing s.token() encode sub as a JSON number.
	// They must keep parsing to the exact user id after this fix.
	const secret = "change-this-secret"
	for _, id := range []uint{1, 5, 42, 9999} {
		tok := mintTokenForTesting(t, id, secret)
		got, err := ParseUserID(tok, secret)
		if err != nil {
			t.Fatalf("id=%d: %v", id, err)
		}
		if got != id {
			t.Errorf("id=%d: got %d (regression — old numeric tokens must still resolve)", id, got)
		}
	}
}
