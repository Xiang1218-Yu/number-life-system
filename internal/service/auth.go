package service

import (
	"encoding/json"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"number-life-system/internal/domain"
	"strconv"
	"time"
)

type AuthService struct {
	DB     *gorm.DB
	Secret string
}
type AuthResult struct {
	Token string      `json:"token"`
	User  domain.User `json:"user"`
}

func (s *AuthService) Register(email, name, plain string) (AuthResult, error) {
	if email == "" || name == "" || len(plain) < 8 {
		return AuthResult{}, errors.New("email、姓名和至少8位密码不能为空")
	}
	var exists domain.User
	if err := s.DB.Where("email = ?", email).First(&exists).Error; err == nil {
		return AuthResult{}, errors.New("邮箱已注册")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return AuthResult{}, err
	}
	user := domain.User{Email: email, Name: name, PasswordHash: string(hash)}
	if err := s.DB.Create(&user).Error; err != nil {
		return AuthResult{}, err
	}
	token, err := s.token(user.ID, user.Email)
	return AuthResult{Token: token, User: user}, err
}
func (s *AuthService) Login(email, plain string) (AuthResult, error) {
	var user domain.User
	if err := s.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return AuthResult{}, errors.New("邮箱或密码错误")
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(plain)) != nil {
		return AuthResult{}, errors.New("邮箱或密码错误")
	}
	token, err := s.token(user.ID, user.Email)
	return AuthResult{Token: token, User: user}, err
}
func (s *AuthService) token(id uint, email string) (string, error) {
	claims := jwt.MapClaims{"sub": id, "email": email, "jti": uuid.NewString(), "exp": time.Now().Add(30 * 24 * time.Hour).Unix()}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.Secret))
}
func ParseUserID(token string, secret string) (uint, error) {
	if token == "" {
		return 0, errors.New("empty token")
	}
	// Use UseJSONNumber so numeric claims survive as json.Number instead of the
	// lossy float64 default; json.Number also reaches our coercion below intact.
	parsed, err := jwt.NewParser(jwt.WithJSONNumber()).Parse(token, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("invalid signing method")
		}
		return []byte(secret), nil
	})
	if err != nil || !parsed.Valid {
		return 0, errors.New("invalid token")
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid claims")
	}
	id, ok := coerceUserID(claims["sub"])
	if !ok || id == 0 {
		return 0, errors.New("invalid user")
	}
	return id, nil
}

// coerceUserID extracts a non-zero user id from a "sub" claim regardless of how
// it was JSON-encoded: as a number (float64, json.Number, or any integer kind)
// or as a decimal string (e.g. tokens minted by other issuers / migrated users).
// It returns the exact id that was signed — never an offset of it.
func coerceUserID(raw any) (uint, bool) {
	switch value := raw.(type) {
	case nil:
		return 0, false
	case float64:
		if value <= 0 || value != float64(uint64(value)) {
			return 0, false
		}
		return uint(value), true
	case json.Number:
		return parseDecimal(value.String())
	case string:
		return parseDecimal(value)
	case uint:
		return value, true
	case uint64:
		return uint(value), true
	case int:
		if value <= 0 {
			return 0, false
		}
		return uint(value), true
	case int64:
		if value <= 0 {
			return 0, false
		}
		return uint(value), true
	}
	return 0, false
}

// parseDecimal parses a base-10 integer string into a uint, rejecting empties,
// signs, and overflow.
func parseDecimal(s string) (uint, bool) {
	if s == "" {
		return 0, false
	}
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, false
	}
	if n == 0 {
		return 0, false
	}
	return uint(n), true
}
