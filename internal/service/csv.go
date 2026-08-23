package service

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"number-life-system/internal/domain"
)

type CSVService struct {
	DB *gorm.DB
}

var accountCSVHeader = []string{
	"platform",
	"username",
	"email",
	"category",
	"registered_at",
	"password_strength",
	"password_changed_at",
	"two_factor_enabled",
	"known_breach",
	"last_login_at",
	"notes",
	"status",
}

func (s *CSVService) ExportAccounts(userID uint, includeArchived bool) ([]byte, error) {
	var rows []domain.Account
	query := s.DB.Where("user_id = ?", userID)
	if !includeArchived {
		query = query.Where("status = ?", "active")
	}
	if err := query.Order("id asc").Find(&rows).Error; err != nil {
		return nil, err
	}
	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)
	if err := writer.Write(accountCSVHeader); err != nil {
		return nil, err
	}
	for _, row := range rows {
		if err := writer.Write(accountCSVRecord(row)); err != nil {
			return nil, err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func accountCSVRecord(row domain.Account) []string {
	return []string{
		row.Platform,
		row.Username,
		row.Email,
		row.Category,
		formatCSVTime(row.RegisteredAt),
		row.PasswordStrength,
		formatCSVTime(row.PasswordChangedAt),
		strconv.FormatBool(row.TwoFactorEnabled),
		strconv.FormatBool(row.KnownBreach),
		formatCSVTime(row.LastLoginAt),
		row.Notes,
		row.Status,
	}
}

func formatCSVTime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.Format(time.RFC3339)
}

func (s *CSVService) ImportAccounts(userID uint, data []byte) (ImportResult, error) {
	reader := csv.NewReader(bytes.NewReader(data))
	reader.FieldsPerRecord = -1
	header, err := reader.Read()
	if err != nil {
		return ImportResult{}, errors.New("CSV 文件为空")
	}
	if err := validateCSVHeader(header); err != nil {
		return ImportResult{}, err
	}
	result := ImportResult{}
	for line := 2; ; line++ {
		record, readErr := reader.Read()
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("第%d行读取失败: %v", line, readErr))
			continue
		}
		if len(record) != len(accountCSVHeader) {
			result.Errors = append(result.Errors, fmt.Sprintf("第%d行字段数不正确", line))
			continue
		}
		input, convertErr := accountCSVInput(record)
		if convertErr != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("第%d行: %v", line, convertErr))
			continue
		}
		row := domain.Account{
			UserID:            userID,
			Platform:          input.Platform,
			Username:          input.Username,
			Email:             input.Email,
			Category:          input.Category,
			RegisteredAt:      input.RegisteredAt,
			PasswordStrength:  input.PasswordStrength,
			PasswordChangedAt: input.PasswordChangedAt,
			TwoFactorEnabled:  input.TwoFactorEnabled,
			KnownBreach:       input.KnownBreach,
			LastLoginAt:       input.LastLoginAt,
			Notes:             input.Notes,
			Status:            normalizeAccountStatus(input.Status),
		}
		if row.PasswordStrength == "" {
			row.PasswordStrength = "weak"
		}
		if err := s.DB.Create(&row).Error; err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("第%d行写入失败: %v", line, err))
			continue
		}
		result.Created++
	}
	return result, nil
}

type ImportResult struct {
	Created int      `json:"created"`
	Errors  []string `json:"errors"`
}

func validateCSVHeader(header []string) error {
	if len(header) != len(accountCSVHeader) {
		return errors.New("CSV 表头不匹配")
	}
	for index, value := range header {
		if strings.TrimSpace(value) != accountCSVHeader[index] {
			return errors.New("CSV 表头不匹配")
		}
	}
	return nil
}

func accountCSVInput(record []string) (AccountInput, error) {
	registeredAt, err := parseCSVTime(record[4])
	if err != nil {
		return AccountInput{}, errors.New("注册时间格式不正确")
	}
	passwordChangedAt, err := parseCSVTime(record[6])
	if err != nil {
		return AccountInput{}, errors.New("密码修改时间格式不正确")
	}
	lastLoginAt, err := parseCSVTime(record[9])
	if err != nil {
		return AccountInput{}, errors.New("登录时间格式不正确")
	}
	twoFactor, err := strconv.ParseBool(record[7])
	if err != nil {
		return AccountInput{}, errors.New("两步验证字段必须是 true 或 false")
	}
	breach, err := strconv.ParseBool(record[8])
	if err != nil {
		return AccountInput{}, errors.New("泄露标记字段必须是 true 或 false")
	}
	input := AccountInput{
		Platform:          strings.TrimSpace(record[0]),
		Username:          strings.TrimSpace(record[1]),
		Email:             strings.TrimSpace(record[2]),
		Category:          strings.TrimSpace(record[3]),
		RegisteredAt:      registeredAt,
		PasswordStrength:  strings.TrimSpace(record[5]),
		PasswordChangedAt: passwordChangedAt,
		TwoFactorEnabled:  twoFactor,
		KnownBreach:       breach,
		LastLoginAt:       lastLoginAt,
		Notes:             strings.TrimSpace(record[10]),
		Status:            strings.TrimSpace(record[11]),
	}
	if err := validateImportedAccount(input); err != nil {
		return AccountInput{}, err
	}
	return input, nil
}

func validateImportedAccount(input AccountInput) error {
	if strings.TrimSpace(input.Platform) == "" {
		return errors.New("平台名称不能为空")
	}
	if strings.TrimSpace(input.Username) == "" {
		return errors.New("用户名不能为空")
	}
	if input.Email != "" && !strings.Contains(input.Email, "@") {
		return errors.New("邮箱格式不正确")
	}
	if input.PasswordStrength != "" && input.PasswordStrength != "weak" && input.PasswordStrength != "medium" && input.PasswordStrength != "strong" {
		return errors.New("密码强度必须是 weak、medium 或 strong")
	}
	return nil
}

func normalizeAccountStatus(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "archived":
		return "active"
	case "active":
		return "active"
	case "":
		return "active"
	default:
		return "active"
	}
}

func parseCSVTime(value string) (*time.Time, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(value))
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}
