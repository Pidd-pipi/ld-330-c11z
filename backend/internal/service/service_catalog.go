package service

import (
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
func (s *CatalogService) Template(in dto.TemplateInput) (*model.RecordTemplate, error) {
	v := &model.RecordTemplate{
		Name:       in.Name,
		RecordType: in.RecordType,
		Versions:   []model.RecordTemplateVersion{{Version: 1, Content: in.Content}},
	}
	if e := s.repo.CreateTemplate(v); e != nil {
		return nil, e
	}
	return s.repo.FindTemplateByID(v.ID)
}

// PublishTemplateVersion 为既有模板追加一个新版本，历史版本保留不动。
func (s *CatalogService) PublishTemplateVersion(templateID uint, in dto.TemplateVersionInput) (*model.RecordTemplate, error) {
	if _, e := s.repo.AddTemplateVersion(templateID, in.Content); e != nil {
		return nil, e
	}
	return s.repo.FindTemplateByID(templateID)
}
func (s *CatalogService) Templates() ([]model.RecordTemplate, error) { return s.repo.ListTemplates() }

// TemplateOptions 返回每个模板的最新一版，供医生书写病历时选择。
func (s *CatalogService) TemplateOptions() ([]dto.TemplateView, error) {
	templates, e := s.repo.ListTemplates()
	if e != nil {
		return nil, e
	}
	views := make([]dto.TemplateView, 0, len(templates))
	for _, t := range templates {
		if len(t.Versions) == 0 {
			continue
		}
		latest := t.Versions[0]
		views = append(views, dto.TemplateView{
			ID:         t.ID,
			Name:       t.Name,
			RecordType: t.RecordType,
			VersionID:  latest.ID,
			Version:    latest.Version,
			Content:    latest.Content,
			UpdatedAt:  latest.CreatedAt,
		})
	}
	return views, nil
}
func (s *CatalogService) Audit(user uint, username, action, resource, detail string) {
	if e := s.repo.CreateAudit(&model.AuditLog{UserID: user, Username: username, Action: action, Resource: resource, Detail: detail}); e != nil {
		s.logger.Error("audit failure", "error", e)
	}
}
func (s *CatalogService) Audits() ([]model.AuditLog, error) { return s.repo.ListAudit() }
