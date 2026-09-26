package service

import (
	"errors"
	"fmt"
	"github.com/blueship581/gbemr/internal/dto"
	"github.com/blueship581/gbemr/internal/model"
	"github.com/blueship581/gbemr/internal/repository"
	"log/slog"
	"time"
)

type RecordService struct {
	repo        repository.RecordRepository
	patients    repository.PatientRepository
	departments repository.DepartmentRepository
	catalog     repository.CatalogRepository
	logger      *slog.Logger
}

func NewRecordService(r repository.RecordRepository, p repository.PatientRepository, d repository.DepartmentRepository, c repository.CatalogRepository, l *slog.Logger) *RecordService {
	return &RecordService{r, p, d, c, l}
}
func (s *RecordService) Create(in dto.RecordInput, doctorID uint) (*model.MedicalRecord, error) {
	if _, e := s.patients.FindByID(in.PatientID); e != nil {
		return nil, fmt.Errorf("validate patient: %w", e)
	}
	if _, e := s.departments.FindByID(in.DepartmentID); e != nil {
		return nil, fmt.Errorf("validate department: %w", e)
	}
	v := &model.MedicalRecord{PatientID: in.PatientID, DoctorID: doctorID, DepartmentID: in.DepartmentID, RecordType: in.RecordType, ChiefComplaint: in.ChiefComplaint, PresentIllness: in.PresentIllness, PastHistory: in.PastHistory, PhysicalExam: in.PhysicalExam, AuxiliaryExam: in.AuxiliaryExam, Diagnosis: in.Diagnosis, TreatmentPlan: in.TreatmentPlan, RichContent: in.RichContent, Status: "draft"}
	if in.TemplateVersionID != nil {
		ver, e := s.catalog.FindTemplateVersionByID(*in.TemplateVersionID)
		if e != nil {
			return nil, fmt.Errorf("validate template version: %w", e)
		}
		if ver.Template == nil {
			return nil, fmt.Errorf("validate template version: template missing")
		}
		// 快照当时选用的模板名称、版本号与内容，归档病历不随模板更新而变化。
		v.TemplateID = &ver.TemplateID
		v.TemplateName = ver.Template.Name
		v.TemplateVersion = ver.Version
		v.TemplateContent = ver.Content
	}
	if e := s.repo.Create(v); e != nil {
		return nil, fmt.Errorf("create record service: %w", e)
	}
	return s.repo.FindByID(v.ID)
}
func (s *RecordService) List(f map[string]string, page, size int) ([]model.MedicalRecord, int64, error) {
	return s.repo.Search(f, page, size)
}
func (s *RecordService) Get(id uint) (*model.MedicalRecord, error) { return s.repo.FindByID(id) }
func (s *RecordService) Review(id, reviewerID uint, archive bool) (*model.MedicalRecord, error) {
	v, e := s.repo.FindByID(id)
	if e != nil {
		return nil, e
	}
	if v.Status == "archived" {
		return nil, errors.New("archived record cannot be reviewed")
	}
	now := time.Now()
	v.ReviewerID = &reviewerID
	v.ReviewedAt = &now
	if archive {
		v.Status = "archived"
		v.ArchivedAt = &now
	} else {
		v.Status = "reviewed"
	}
	if e = s.repo.Update(v); e != nil {
		return nil, e
	}
	return s.repo.FindByID(id)
}
func (s *RecordService) RequestChange(id, applicant uint, reason string) error {
	v, e := s.repo.FindByID(id)
	if e != nil {
		return e
	}
	if v.Status != "archived" {
		return errors.New("only archived records need a change request")
	}
	return s.repo.CreateChangeRequest(&model.RecordChangeRequest{MedicalRecordID: id, ApplicantID: applicant, Reason: reason, Status: "pending"})
}
