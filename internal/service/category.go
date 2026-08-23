package service

import (
	"errors"
	"strings"

	"gorm.io/gorm"
	"number-life-system/internal/domain"
)

type CategoryService struct {
	DB *gorm.DB
}

type CategoryInput struct {
	Name  string `json:"name" binding:"required"`
	Color string `json:"color"`
}

func (s *CategoryService) List() ([]domain.AccountCategory, error) {
	var rows []domain.AccountCategory
	err := s.DB.Order("name asc").Find(&rows).Error
	return rows, err
}

func (s *CategoryService) Get(id uint) (domain.AccountCategory, error) {
	var row domain.AccountCategory
	err := s.DB.First(&row, id).Error
	return row, err
}

func (s *CategoryService) Create(input CategoryInput) (domain.AccountCategory, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return domain.AccountCategory{}, errors.New("分类名称不能为空")
	}
	if len([]rune(name)) > 60 {
		return domain.AccountCategory{}, errors.New("分类名称不能超过60个字符")
	}
	if input.Color == "" {
		input.Color = "#64748b"
	}
	row := domain.AccountCategory{Name: name, Color: input.Color}
	if err := s.DB.Create(&row).Error; err != nil {
		return domain.AccountCategory{}, errors.New("分类名称已存在")
	}
	return row, nil
}

func (s *CategoryService) Update(id uint, input CategoryInput) (domain.AccountCategory, error) {
	row, err := s.Get(id)
	if err != nil {
		return row, err
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return row, errors.New("分类名称不能为空")
	}
	if input.Color == "" {
		input.Color = row.Color
	}
	row.Name = name
	row.Color = input.Color
	if err := s.DB.Save(&row).Error; err != nil {
		return row, errors.New("分类名称已存在")
	}
	return row, nil
}

func (s *CategoryService) Delete(id uint) error {
	row, err := s.Get(id)
	if err != nil {
		return errors.New("分类不存在")
	}
	var count int64
	if err := s.DB.Model(&domain.Account{}).Where("category = ?", row.Name).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("分类仍被账户使用，不能删除")
	}
	result := s.DB.Delete(&domain.AccountCategory{}, id)
	if result.RowsAffected == 0 {
		return errors.New("分类不存在")
	}
	return result.Error
}
