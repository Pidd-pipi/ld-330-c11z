package repository

import (
	"errors"
	"fmt"
	"github.com/blueship581/gbemr/internal/model"
	"gorm.io/gorm"
	"time"
)

type CatalogRepository interface {
	CreateDrug(*model.Drug) error
	ListDrugs() ([]model.Drug, error)
	CreateDiagnosis(*model.DiagnosisCode) error
	ListDiagnoses() ([]model.DiagnosisCode, error)
	CreateTemplate(*model.RecordTemplate) error
	AddTemplateVersion(templateID uint, content string) (*model.RecordTemplateVersion, error)
	FindTemplateByID(uint) (*model.RecordTemplate, error)
	FindTemplateVersionByID(uint) (*model.RecordTemplateVersion, error)
	ListTemplates() ([]model.RecordTemplate, error)
	CreateAudit(*model.AuditLog) error
	ListAudit() ([]model.AuditLog, error)
}
type catalogRepository struct{ db *gorm.DB }

func NewCatalogRepository(db *gorm.DB) CatalogRepository    { return &catalogRepository{db} }
func (r *catalogRepository) CreateDrug(v *model.Drug) error { return r.db.Create(v).Error }
func (r *catalogRepository) ListDrugs() ([]model.Drug, error) {
	var v []model.Drug
	return v, r.db.Order("name").Find(&v).Error
}
func (r *catalogRepository) CreateDiagnosis(v *model.DiagnosisCode) error {
	return r.db.Create(v).Error
}
func (r *catalogRepository) ListDiagnoses() ([]model.DiagnosisCode, error) {
	var v []model.DiagnosisCode
	return v, r.db.Order("code").Find(&v).Error
}
func (r *catalogRepository) CreateTemplate(v *model.RecordTemplate) error {
	return r.db.Create(v).Error
}

// AddTemplateVersion 在事务内计算下一个版本号并插入新版本，历史版本行保持不动；
// 版本号上的复合唯一索引保证并发更新时也不会出现重复版本。
func (r *catalogRepository) AddTemplateVersion(templateID uint, content string) (*model.RecordTemplateVersion, error) {
	var v model.RecordTemplateVersion
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var t model.RecordTemplate
		if err := tx.First(&t, templateID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return fmt.Errorf("find template: %w", err)
		}
		var next int
		if err := tx.Model(&model.RecordTemplateVersion{}).Where("template_id = ?", templateID).
			Select("COALESCE(MAX(version), 0) + 1").Scan(&next).Error; err != nil {
			return fmt.Errorf("compute next template version: %w", err)
		}
		v = model.RecordTemplateVersion{TemplateID: templateID, Version: next, Content: content}
		if err := tx.Create(&v).Error; err != nil {
			return fmt.Errorf("create template version: %w", err)
		}
		if err := tx.Model(&model.RecordTemplate{}).Where("id = ?", templateID).Update("updated_at", time.Now()).Error; err != nil {
			return fmt.Errorf("touch template: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &v, nil
}
func (r *catalogRepository) FindTemplateByID(id uint) (*model.RecordTemplate, error) {
	var v model.RecordTemplate
	if err := r.db.Preload("Versions", func(db *gorm.DB) *gorm.DB { return db.Order("version desc") }).
		First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find template: %w", err)
	}
	return &v, nil
}
func (r *catalogRepository) FindTemplateVersionByID(id uint) (*model.RecordTemplateVersion, error) {
	var v model.RecordTemplateVersion
	if err := r.db.Preload("Template").First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find template version: %w", err)
	}
	return &v, nil
}
func (r *catalogRepository) ListTemplates() ([]model.RecordTemplate, error) {
	var v []model.RecordTemplate
	return v, r.db.Preload("Versions", func(db *gorm.DB) *gorm.DB { return db.Order("version desc") }).
		Order("name").Find(&v).Error
}
func (r *catalogRepository) CreateAudit(v *model.AuditLog) error {
	if err := r.db.Create(v).Error; err != nil {
		return fmt.Errorf("create audit log: %w", err)
	}
	return nil
}
func (r *catalogRepository) ListAudit() ([]model.AuditLog, error) {
	var v []model.AuditLog
	if err := r.db.Order("created_at desc").Limit(200).Find(&v).Error; err != nil {
		return nil, fmt.Errorf("list audit logs: %w", err)
	}
	return v, nil
}
