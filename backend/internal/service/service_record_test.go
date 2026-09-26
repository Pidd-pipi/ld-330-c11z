package service

import (
	"errors"
	"github.com/blueship581/gbemr/internal/dto"
	"github.com/blueship581/gbemr/internal/model"
	"github.com/blueship581/gbemr/internal/repository"
	"log/slog"
	"testing"
)

type recordRepoFake struct{ created *model.MedicalRecord }

func (f *recordRepoFake) Create(v *model.MedicalRecord) error { v.ID = 1; f.created = v; return nil }
func (f *recordRepoFake) FindByID(uint) (*model.MedicalRecord, error) {
	return f.created, nil
}
func (f *recordRepoFake) Search(map[string]string, int, int) ([]model.MedicalRecord, int64, error) {
	return nil, 0, nil
}
func (f *recordRepoFake) Update(*model.MedicalRecord) error                    { return nil }
func (f *recordRepoFake) CreateChangeRequest(*model.RecordChangeRequest) error { return nil }

type departmentRepoFake struct{}

func (departmentRepoFake) Create(*model.Department) error { return nil }
func (departmentRepoFake) List() ([]model.Department, error) {
	return nil, nil
}
func (departmentRepoFake) FindByID(id uint) (*model.Department, error) {
	return &model.Department{ID: id, Name: "内科"}, nil
}

func TestRecordServiceCreateSnapshotsTemplate(t *testing.T) {
	catalog := newCatalogRepoFake()
	tmpl := &model.RecordTemplate{Name: "门诊初诊模板", RecordType: "outpatient",
		Versions: []model.RecordTemplateVersion{{Version: 1, Content: "主诉："}}}
	if e := catalog.CreateTemplate(tmpl); e != nil {
		t.Fatal(e)
	}
	versionID := tmpl.Versions[0].ID
	repo := &recordRepoFake{}
	s := NewRecordService(repo, &patientRepoFake{}, departmentRepoFake{}, catalog, slog.Default())
	base := dto.RecordInput{PatientID: 1, DepartmentID: 1, RecordType: "outpatient", ChiefComplaint: "头痛", Diagnosis: "感冒"}
	cases := []struct {
		name         string
		in           dto.RecordInput
		wantTemplate bool
	}{
		{"with template version snapshots name version content", dto.RecordInput{PatientID: 1, DepartmentID: 1, RecordType: "outpatient", ChiefComplaint: "头痛", Diagnosis: "感冒", TemplateVersionID: &versionID}, true},
		{"without template keeps snapshot empty", base, false},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			rec, e := s.Create(tt.in, 7)
			if e != nil {
				t.Fatal(e)
			}
			if tt.wantTemplate {
				if rec.TemplateID == nil || *rec.TemplateID != tmpl.ID {
					t.Fatalf("template id not linked: %#v", rec.TemplateID)
				}
				if rec.TemplateName != "门诊初诊模板" || rec.TemplateVersion != 1 || rec.TemplateContent != "主诉：" {
					t.Fatalf("snapshot mismatch: %#v", rec)
				}
			} else if rec.TemplateID != nil || rec.TemplateName != "" || rec.TemplateVersion != 0 || rec.TemplateContent != "" {
				t.Fatalf("unexpected template snapshot: %#v", rec)
			}
		})
	}
}

func TestRecordServiceCreateRejectsUnknownTemplateVersion(t *testing.T) {
	s := NewRecordService(&recordRepoFake{}, &patientRepoFake{}, departmentRepoFake{}, newCatalogRepoFake(), slog.Default())
	missing := uint(42)
	_, e := s.Create(dto.RecordInput{PatientID: 1, DepartmentID: 1, RecordType: "outpatient", ChiefComplaint: "头痛", Diagnosis: "感冒", TemplateVersionID: &missing}, 7)
	if e == nil {
		t.Fatal("expected error for unknown template version")
	}
	if !errors.Is(e, repository.ErrNotFound) {
		t.Fatalf("expected not found error, got %v", e)
	}
}
