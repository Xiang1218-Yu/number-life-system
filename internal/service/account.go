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
	Platform          string       `json:"platform" binding:"required"`
	Username          string       `json:"username" binding:"required"`
	Email             string       `json:"email"`
	Category          string       `json:"category" binding:"required"`
	RegisteredAt      FlexibleTime `json:"registered_at"`
	Password          string       `json:"password"`
	PasswordStrength  string       `json:"password_strength"`
	PasswordChangedAt FlexibleTime `json:"password_changed_at"`
	TwoFactorEnabled  bool         `json:"two_factor_enabled"`
	KnownBreach       bool         `json:"known_breach"`
	LastLoginAt       FlexibleTime `json:"last_login_at"`
	Notes             string       `json:"notes"`
	Status            string       `json:"status"`
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
	row := domain.Account{UserID: userID, Platform: input.Platform, Username: input.Username, Email: input.Email, Category: input.Category, RegisteredAt: input.RegisteredAt.Time(), PasswordChangedAt: input.PasswordChangedAt.Time(), TwoFactorEnabled: input.TwoFactorEnabled, KnownBreach: input.KnownBreach, LastLoginAt: input.LastLoginAt.Time(), Notes: input.Notes, Status: "active", PasswordStrength: password.Strength(input.Password)}
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
	row.Platform, row.Username, row.Email, row.Category, row.Notes = input.Platform, input.Username, input.Email, input.Category, input.Notes
	row.TwoFactorEnabled, row.KnownBreach = input.TwoFactorEnabled, input.KnownBreach
	// Time fields are optional: a provided value (date-only or RFC3339) overwrites
	// the stored value; an omitted/empty value leaves the stored value intact so
	// re-saving the same record cannot wipe a previously recorded login or
	// password-change time and silently flip the security score. Validation has
	// already rejected future timestamps, so the stored value equals the input.
	row.RegisteredAt = mergeAccountTime(row.RegisteredAt, input.RegisteredAt.Time())
	row.PasswordChangedAt = mergeAccountTime(row.PasswordChangedAt, input.PasswordChangedAt.Time())
	row.LastLoginAt = mergeAccountTime(row.LastLoginAt, input.LastLoginAt.Time())
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

// mergeAccountTime applies a user-supplied time on top of an existing stored
// value. A nil input means "not provided" and preserves the stored value, so an
// update that omits an optional time field does not clear it. A non-nil input
// replaces the stored value with exactly what the user entered.
func mergeAccountTime(stored, input *time.Time) *time.Time {
	if input == nil {
		return stored
	}
	return input
}
func (s *AccountService) Delete(userID, id uint) error {
	result := s.DB.Model(&domain.Account{}).Where("user_id = ? AND id = ?", userID, id).Updates(map[string]any{"status": "archived", "archived_at": time.Now(), "archive_reason": "手动归档"})
	if result.RowsAffected == 0 {
		return errors.New("账户不存在")
	}
	return result.Error
}
