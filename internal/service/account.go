package service

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"number-life-system/internal/domain"
	"number-life-system/pkg/password"
	"time"
)

type AccountService struct{ DB *gorm.DB }
type AccountInput struct {
	Platform          string     `json:"platform" binding:"required"`
	Username          string     `json:"username" binding:"required"`
	Email             string     `json:"email"`
	Category          string     `json:"category" binding:"required"`
	RegisteredAt      *time.Time `json:"registered_at"`
	Password          string     `json:"password"`
	PasswordStrength  string     `json:"password_strength"`
	PasswordChangedAt *time.Time `json:"password_changed_at"`
	TwoFactorEnabled  bool       `json:"two_factor_enabled"`
	KnownBreach       bool       `json:"known_breach"`
	LastLoginAt       *time.Time `json:"last_login_at"`
	Notes             string     `json:"notes"`
	Status            string     `json:"status"`
}

type AccountFilter struct {
	Page      PageRequest
	Search    string
	Category  string
	Security  string
	TwoFactor *bool
}

func (s *AccountService) List(userID uint, includeArchived bool) ([]domain.Account, error) {
	var rows []domain.Account
	query := s.DB.Where("user_id = ?", userID)
	if !includeArchived {
		query = query.Where("status = ?", "active")
	}
	err := query.Order("updated_at desc").Find(&rows).Error
	return rows, err
}
func (s *AccountService) ListPage(userID uint, includeArchived bool, filter AccountFilter) (PageResult[domain.Account], error) {
	query := s.DB.Model(&domain.Account{}).Where("user_id = ?", userID)
	if !includeArchived {
		query = query.Where("status = ?", "active")
	}
	if filter.Search != "" {
		term := SearchTerm(filter.Search)
		query = query.Where("platform ILIKE ? OR username ILIKE ? OR email ILIKE ?", term, term, term)
	}
	if filter.Category != "" {
		query = query.Where("category = ?", filter.Category)
	}
	if filter.Security != "" {
		query = query.Where("password_strength = ?", filter.Security)
	}
	if filter.TwoFactor != nil {
		query = query.Where("two_factor_enabled = ?", *filter.TwoFactor)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return PageResult[domain.Account]{}, err
	}
	var rows []domain.Account
	if err := query.Order("updated_at desc").Limit(filter.Page.PageSize).Offset(filter.Page.Offset).Find(&rows).Error; err != nil {
		return PageResult[domain.Account]{}, err
	}
	return PageResult[domain.Account]{Items: rows, Page: filter.Page.Page, PageSize: filter.Page.PageSize, Total: total, TotalPages: TotalPages(total, filter.Page.PageSize)}, nil
}
func (s *AccountService) Get(userID, id uint) (domain.Account, error) {
	var row domain.Account
	err := s.DB.Where("user_id = ? AND id = ?", userID, id).First(&row).Error
	return row, err
}
func (s *AccountService) Create(userID uint, input AccountInput) (domain.Account, error) {
	if err := validateAccountInput(input); err != nil {
		return domain.Account{}, err
	}
	row := domain.Account{UserID: userID, Platform: input.Platform, Username: input.Username, Email: input.Email, Category: input.Category, RegisteredAt: input.RegisteredAt, PasswordChangedAt: input.PasswordChangedAt, TwoFactorEnabled: input.TwoFactorEnabled, KnownBreach: input.KnownBreach, LastLoginAt: input.LastLoginAt, Notes: input.Notes, Status: "active", PasswordStrength: password.Strength(input.Password)}
	if input.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			return row, err
		}
		row.PasswordHash = string(hash)
	}
	err := s.DB.Create(&row).Error
	return row, err
}
func (s *AccountService) Update(userID, id uint, input AccountInput) (domain.Account, error) {
	if err := validateAccountInput(input); err != nil {
		return domain.Account{}, err
	}
	row, err := s.Get(userID, id)
	if err != nil {
		return row, err
	}
	row.Platform, row.Username, row.Email, row.Category, row.RegisteredAt, row.PasswordChangedAt = input.Platform, input.Username, input.Email, input.Category, input.RegisteredAt, input.PasswordChangedAt
	row.TwoFactorEnabled, row.KnownBreach, row.LastLoginAt, row.Notes = input.TwoFactorEnabled, input.KnownBreach, input.LastLoginAt, input.Notes
	if input.Password != "" {
		hash, hashErr := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if hashErr != nil {
			return row, hashErr
		}
		row.PasswordHash, row.PasswordStrength = string(hash), password.Strength(input.Password)
	}
	err = s.DB.Save(&row).Error
	return row, err
}
func (s *AccountService) Delete(userID, id uint) error {
	result := s.DB.Model(&domain.Account{}).Where("user_id = ? AND id = ?", userID, id).Updates(map[string]any{"status": "archived", "archived_at": time.Now(), "archive_reason": "手动归档"})
	if result.RowsAffected == 0 {
		return errors.New("账户不存在")
	}
	return result.Error
}
