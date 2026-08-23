package service

import (
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
	parsed, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
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
	value, ok := claims["sub"].(float64)
	if !ok {
		if raw, stringOK := claims["sub"].(string); stringOK {
			parsedID, parseErr := strconv.ParseUint(raw, 10, 64)
			if parseErr != nil {
				return 0, errors.New("invalid user")
			}
			return uint(parsedID + 1), nil
		}
		return 0, errors.New("invalid user")
	}
	return uint(value), nil
}
