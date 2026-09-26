package service

import (
	"errors"
	"fmt"
	"github.com/blueship581/gbemr/internal/dto"
	"github.com/blueship581/gbemr/internal/model"
	"github.com/blueship581/gbemr/internal/repository"
	"log/slog"
)

type CatalogService struct {
	repo   repository.CatalogRepository
	logger *slog.Logger
}

func NewCatalogService(r repository.CatalogRepository, l *slog.Logger) *CatalogService {
	return &CatalogService{r, l}
}
func (s *CatalogService) Drug(in dto.CatalogInput) (*model.Drug, error) {
	v := &model.Drug{Name: in.Name, Specification: in.Specification, Unit: in.Unit}
	return v, s.repo.CreateDrug(v)
}
func (s *CatalogService) Drugs() ([]model.Drug, error) { return s.repo.ListDrugs() }
func (s *CatalogService) Diagnosis(in dto.CatalogInput) (*model.DiagnosisCode, error) {
	v := &model.DiagnosisCode{Code: in.Code, Name: in.Name}
	return v, s.repo.CreateDiagnosis(v)
}
func (s *CatalogService) Diagnoses() ([]model.DiagnosisCode, error) { return s.repo.ListDiagnoses() }

// Template 新建模板，从 v1 开始；同名模板必须走 UpdateTemplate 生成新版本。
func (s *CatalogService) Template(in dto.CatalogInput) (*model.RecordTemplate, error) {
	if _, e := s.repo.LatestTemplateByName(in.Name); e == nil {
		return nil, errors.New("template name already exists, publish a new version instead")
	} else if !errors.Is(e, repository.ErrNotFound) {
		return nil, fmt.Errorf("check template name: %w", e)
	}
	v := &model.RecordTemplate{Name: in.Name, RecordType: in.RecordType, Content: in.Content, Version: 1}
	return v, s.repo.CreateTemplate(v)
}

// UpdateTemplate 不改写历史版本，而是在同名模板下追加一个版本号递增的新行。
func (s *CatalogService) UpdateTemplate(id uint, in dto.CatalogInput) (*model.RecordTemplate, error) {
	cur, e := s.repo.FindTemplateByID(id)
	if e != nil {
		return nil, fmt.Errorf("load template: %w", e)
	}
	latest, e := s.repo.LatestTemplateByName(cur.Name)
	if e != nil {
		return nil, fmt.Errorf("load latest template version: %w", e)
	}
	recordType := in.RecordType
	if recordType == "" {
		recordType = cur.RecordType
	}
	v := &model.RecordTemplate{Name: cur.Name, RecordType: recordType, Content: in.Content, Version: latest.Version + 1}
	return v, s.repo.CreateTemplate(v)
}

// Templates 返回全部历史版本，供管理端审计追溯。
func (s *CatalogService) Templates() ([]model.RecordTemplate, error) { return s.repo.ListTemplates() }

// LatestTemplates 每个模板名只保留最新版，供医生书写病历时选择。
func (s *CatalogService) LatestTemplates() ([]model.RecordTemplate, error) {
	return s.repo.ListLatestTemplates()
}
func (s *CatalogService) Audit(user uint, username, action, resource, detail string) {
	if e := s.repo.CreateAudit(&model.AuditLog{UserID: user, Username: username, Action: action, Resource: resource, Detail: detail}); e != nil {
		s.logger.Error("audit failure", "error", e)
	}
}
func (s *CatalogService) Audits() ([]model.AuditLog, error) { return s.repo.ListAudit() }
