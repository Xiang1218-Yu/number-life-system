package service

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestV6StringSubjectKeepsUserIdentity(t *testing.T) {
	const secret = "v6-secret"
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "42", "exp": int64(4102444800)}).SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	id, err := ParseUserID(token, secret)
	if err != nil {
		t.Fatal(err)
	}
	if id != 42 {
		t.Fatalf("parsed user id = %d, want 42", id)
	}
}
