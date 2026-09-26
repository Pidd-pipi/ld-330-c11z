package repository

import (
	"github.com/blueship581/gbemr/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"testing"
)

func TestCatalogRepositoryTemplateVersions(t *testing.T) {
	db, e := gorm.Open(sqlite.Open("file:catalogrepo?mode=memory&cache=shared"), &gorm.Config{})
	if e != nil {
		t.Fatal(e)
	}
	if e = db.AutoMigrate(&model.RecordTemplate{}, &model.RecordTemplateVersion{}); e != nil {
		t.Fatal(e)
	}
	repo := NewCatalogRepository(db)
	tmpl := &model.RecordTemplate{Name: "门诊初诊模板", RecordType: "outpatient",
		Versions: []model.RecordTemplateVersion{{Version: 1, Content: "主诉："}}}
	if e = repo.CreateTemplate(tmpl); e != nil {
		t.Fatal(e)
	}
	if tmpl.Versions[0].ID == 0 || tmpl.Versions[0].TemplateID != tmpl.ID {
		t.Fatalf("initial version not persisted with template: %#v", tmpl.Versions)
	}
	v2, e := repo.AddTemplateVersion(tmpl.ID, "主诉：\n现病史：")
	if e != nil {
		t.Fatal(e)
	}
	if v2.Version != 2 {
		t.Fatalf("expected version 2, got %d", v2.Version)
	}
	v3, e := repo.AddTemplateVersion(tmpl.ID, "主诉：\n现病史：\n既往史：")
	if e != nil {
		t.Fatal(e)
	}
	if v3.Version != 3 {
		t.Fatalf("expected version 3, got %d", v3.Version)
	}
	if _, e = repo.AddTemplateVersion(999, "x"); e != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", e)
	}
	full, e := repo.FindTemplateByID(tmpl.ID)
	if e != nil {
		t.Fatal(e)
	}
	if len(full.Versions) != 3 || full.Versions[0].Version != 3 {
		t.Fatalf("expected 3 versions latest first, got %#v", full.Versions)
	}
	if full.Versions[2].Content != "主诉：" {
		t.Fatalf("oldest version content changed: %#v", full.Versions[2])
	}
	ver, e := repo.FindTemplateVersionByID(full.Versions[2].ID)
	if e != nil {
		t.Fatal(e)
	}
	if ver.Template == nil || ver.Template.Name != "门诊初诊模板" {
		t.Fatalf("version must preload template name: %#v", ver)
	}
	list, e := repo.ListTemplates()
	if e != nil {
		t.Fatal(e)
	}
	if len(list) != 1 || len(list[0].Versions) != 3 {
		t.Fatalf("list templates with versions: %#v", list)
	}
}
