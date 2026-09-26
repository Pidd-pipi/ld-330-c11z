package service

import (
	"github.com/blueship581/gbemr/internal/dto"
	"github.com/blueship581/gbemr/internal/model"
	"github.com/blueship581/gbemr/internal/repository"
	"log/slog"
	"testing"
	"time"
)

// catalogRepoFake 内存模拟版本化模板仓储，行为与真实仓储一致：
// 每次 AddTemplateVersion 追加一个递增版本，历史版本保留。
type catalogRepoFake struct {
	templates map[uint]*model.RecordTemplate
	nextID    uint
}

func newCatalogRepoFake() *catalogRepoFake {
	return &catalogRepoFake{templates: map[uint]*model.RecordTemplate{}, nextID: 1}
}
func (f *catalogRepoFake) CreateDrug(*model.Drug) error               { return nil }
func (f *catalogRepoFake) ListDrugs() ([]model.Drug, error)           { return nil, nil }
func (f *catalogRepoFake) CreateDiagnosis(*model.DiagnosisCode) error { return nil }
func (f *catalogRepoFake) ListDiagnoses() ([]model.DiagnosisCode, error) {
	return nil, nil
}
func (f *catalogRepoFake) CreateTemplate(v *model.RecordTemplate) error {
	v.ID = f.nextID
	f.nextID++
	for i := range v.Versions {
		v.Versions[i].ID = v.ID*100 + uint(i)
		v.Versions[i].TemplateID = v.ID
		v.Versions[i].CreatedAt = time.Now()
	}
	f.templates[v.ID] = v
	return nil
}
func (f *catalogRepoFake) AddTemplateVersion(templateID uint, content string) (*model.RecordTemplateVersion, error) {
	t, ok := f.templates[templateID]
	if !ok {
		return nil, repository.ErrNotFound
	}
	v := &model.RecordTemplateVersion{ID: uint(len(t.Versions)) + 1, TemplateID: templateID, Version: len(t.Versions) + 1, Content: content, CreatedAt: time.Now()}
	t.Versions = append([]model.RecordTemplateVersion{*v}, t.Versions...)
	return v, nil
}
func (f *catalogRepoFake) FindTemplateByID(id uint) (*model.RecordTemplate, error) {
	t, ok := f.templates[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return t, nil
}
func (f *catalogRepoFake) FindTemplateVersionByID(id uint) (*model.RecordTemplateVersion, error) {
	for _, t := range f.templates {
		for _, v := range t.Versions {
			if v.ID == id {
				ver := v
				ver.Template = t
				return &ver, nil
			}
		}
	}
	return nil, repository.ErrNotFound
}
func (f *catalogRepoFake) ListTemplates() ([]model.RecordTemplate, error) {
	out := make([]model.RecordTemplate, 0, len(f.templates))
	for _, t := range f.templates {
		out = append(out, *t)
	}
	return out, nil
}
func (f *catalogRepoFake) CreateAudit(*model.AuditLog) error    { return nil }
func (f *catalogRepoFake) ListAudit() ([]model.AuditLog, error) { return nil, nil }

func TestCatalogServiceTemplateVersioning(t *testing.T) {
	s := NewCatalogService(newCatalogRepoFake(), slog.Default())
	tmpl, e := s.Template(dto.TemplateInput{Name: "门诊初诊模板", RecordType: "outpatient", Content: "主诉：\n现病史："})
	if e != nil {
		t.Fatal(e)
	}
	if len(tmpl.Versions) != 1 || tmpl.Versions[0].Version != 1 {
		t.Fatalf("expected initial version 1, got %#v", tmpl.Versions)
	}
	updated, e := s.PublishTemplateVersion(tmpl.ID, dto.TemplateVersionInput{Content: "主诉：\n现病史：\n既往史："})
	if e != nil {
		t.Fatal(e)
	}
	if len(updated.Versions) != 2 {
		t.Fatalf("expected 2 versions retained, got %d", len(updated.Versions))
	}
	if updated.Versions[0].Version != 2 || updated.Versions[1].Version != 1 {
		t.Fatalf("versions not ordered latest first: %#v", updated.Versions)
	}
	if updated.Versions[1].Content != "主诉：\n现病史：" {
		t.Fatalf("historical version content changed: %#v", updated.Versions[1])
	}
	if _, e = s.PublishTemplateVersion(999, dto.TemplateVersionInput{Content: "x"}); e == nil {
		t.Fatal("expected not found error for unknown template")
	}
}

func TestCatalogServiceTemplateOptions(t *testing.T) {
	s := NewCatalogService(newCatalogRepoFake(), slog.Default())
	tmpl, e := s.Template(dto.TemplateInput{Name: "住院病历模板", RecordType: "inpatient", Content: "旧内容"})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.PublishTemplateVersion(tmpl.ID, dto.TemplateVersionInput{Content: "新内容"}); e != nil {
		t.Fatal(e)
	}
	views, e := s.TemplateOptions()
	if e != nil {
		t.Fatal(e)
	}
	if len(views) != 1 {
		t.Fatalf("expected 1 option, got %d", len(views))
	}
	v := views[0]
	if v.Version != 2 || v.Content != "新内容" || v.Name != "住院病历模板" {
		t.Fatalf("option does not expose latest version: %#v", v)
	}
	if v.VersionID == 0 {
		t.Fatal("option must carry version id for record snapshot")
	}
}
