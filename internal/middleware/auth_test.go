package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func mint(sub any, secret string) string {
	c := jwt.MapClaims{"sub": sub, "exp": time.Now().Add(30 * 24 * time.Hour).Unix()}
	t, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString([]byte(secret))
	return t
}

func TestAuth_AllowsValidStringSubToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const secret = "s"
	var captured uint
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Next() })
	r.GET("/p", Auth(secret), func(c *gin.Context) {
		captured = UserID(c)
		c.Status(http.StatusOK)
	})
	// Migrated token where sub is a string "5" — previously the middleware either
	// rejected (if type missed) or set user_id=6. Now it must set user_id=5 and allow.
	req := httptest.NewRequest(http.MethodGet, "/p", nil)
	req.Header.Set("Authorization", "Bearer "+mint("5", secret))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	if captured != 5 {
		t.Fatalf("expected user_id=5, got %d", captured)
	}

	// Rejected: missing header
	req2 := httptest.NewRequest(http.MethodGet, "/p", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for missing header, got %d", w2.Code)
	}
	if !strings.Contains(w2.Body.String(), "需要登录") {
		t.Fatalf("expected 需要登录, got %s", w2.Body.String())
	}
}
