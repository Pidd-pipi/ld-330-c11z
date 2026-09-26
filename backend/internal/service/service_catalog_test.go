package service

import (
	"github.com/blueship581/gbemr/internal/dto"
	"github.com/blueship581/gbemr/internal/model"
	"github.com/blueship581/gbemr/internal/repository"
	"log/slog"
	"testing"
)

type catalogRepoFake struct {
	versions []model.RecordTemplate
	nextID   uint
}

func (f *catalogRepoFake) CreateDrug(*model.Drug) error               { return nil }
func (f *catalogRepoFake) ListDrugs() ([]model.Drug, error)           { return nil, nil }
func (f *catalogRepoFake) CreateDiagnosis(*model.DiagnosisCode) error { return nil }
func (f *catalogRepoFake) ListDiagnoses() ([]model.DiagnosisCode, error) {
	return nil, nil
}
func (f *catalogRepoFake) CreateTemplate(v *model.RecordTemplate) error {
	f.nextID++
	v.ID = f.nextID
	f.versions = append(f.versions, *v)
	return nil
}
func (f *catalogRepoFake) FindTemplateByID(id uint) (*model.RecordTemplate, error) {
	for i := range f.versions {
		if f.versions[i].ID == id {
			return &f.versions[i], nil
		}
	}
	return nil, repository.ErrNotFound
}
func (f *catalogRepoFake) LatestTemplateByName(name string) (*model.RecordTemplate, error) {
	var latest *model.RecordTemplate
	for i := range f.versions {
		if f.versions[i].Name == name && (latest == nil || f.versions[i].Version > latest.Version) {
			latest = &f.versions[i]
		}
	}
	if latest == nil {
		return nil, repository.ErrNotFound
	}
	return latest, nil
}
func (f *catalogRepoFake) ListTemplates() ([]model.RecordTemplate, error) {
	return f.versions, nil
}
func (f *catalogRepoFake) ListLatestTemplates() ([]model.RecordTemplate, error) {
	latest := map[string]model.RecordTemplate{}
	for _, v := range f.versions {
		if cur, ok := latest[v.Name]; !ok || v.Version > cur.Version {
			latest[v.Name] = v
		}
	}
	out := make([]model.RecordTemplate, 0, len(latest))
	for _, v := range latest {
		out = append(out, v)
	}
	return out, nil
}
func (f *catalogRepoFake) CreateAudit(*model.AuditLog) error    { return nil }
func (f *catalogRepoFake) ListAudit() ([]model.AuditLog, error) { return nil, nil }

func TestCatalogServiceTemplateVersioning(t *testing.T) {
	cases := []struct {
		name        string
		run         func(s *CatalogService) error
		wantVersion int
		wantErr     bool
	}{
		{"first template starts at v1", func(s *CatalogService) error {
			v, e := s.Template(dto.CatalogInput{Name: "门诊初诊模板", RecordType: "outpatient", Content: "主诉："})
			if e == nil && v.Version != 1 {
				t.Fatalf("first version = %d, want 1", v.Version)
			}
			return e
		}, 1, false},
		{"duplicate name rejected on create", func(s *CatalogService) error {
			_, e := s.Template(dto.CatalogInput{Name: "门诊初诊模板", RecordType: "outpatient", Content: "其他内容"})
			return e
		}, 0, true},
		{"update publishes incremented version", func(s *CatalogService) error {
			v, e := s.UpdateTemplate(1, dto.CatalogInput{Content: "主诉：\n现病史："})
			if e == nil && v.Version != 2 {
				t.Fatalf("updated version = %d, want 2", v.Version)
			}
			return e
		}, 2, false},
		{"update missing template fails", func(s *CatalogService) error {
			_, e := s.UpdateTemplate(999, dto.CatalogInput{Content: "x"})
			return e
		}, 0, true},
	}
	f := &catalogRepoFake{}
	s := NewCatalogService(f, slog.Default())
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			e := tt.run(s)
			if tt.wantErr && e == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && e != nil {
				t.Fatalf("unexpected error: %v", e)
			}
		})
	}
	latest, e := s.LatestTemplates()
	if e != nil {
		t.Fatal(e)
	}
	if len(latest) != 1 || latest[0].Version != 2 || latest[0].Content != "主诉：\n现病史：" {
		t.Fatalf("latest templates = %#v, want single v2 template", latest)
	}
	all, e := s.Templates()
	if e != nil {
		t.Fatal(e)
	}
	if len(all) != 2 {
		t.Fatalf("history length = %d, want 2 versions kept", len(all))
	}
}
