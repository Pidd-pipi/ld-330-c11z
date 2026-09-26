package service

import (
	"github.com/blueship581/gbemr/internal/dto"
	"github.com/blueship581/gbemr/internal/model"
	"github.com/blueship581/gbemr/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log/slog"
	"testing"
)

// TestTemplateVersioningFlow 用真实数据库走通完整链路：
// 模板更新产生新版本，医生只看到最新版，病历保存时快照当时版本，归档后不随模板变化。
func TestTemplateVersioningFlow(t *testing.T) {
	db, e := gorm.Open(sqlite.Open("file:templateflow?mode=memory&cache=shared"), &gorm.Config{})
	if e != nil {
		t.Fatal(e)
	}
	if e = db.AutoMigrate(&model.Department{}, &model.User{}, &model.Patient{}, &model.MedicalRecord{}, &model.RecordTemplate{}, &model.RecordTemplateVersion{}); e != nil {
		t.Fatal(e)
	}
	catalogRepo := repository.NewCatalogRepository(db)
	catalogSvc := NewCatalogService(catalogRepo, slog.Default())
	recordSvc := NewRecordService(repository.NewRecordRepository(db), repository.NewPatientRepository(db), repository.NewDepartmentRepository(db), catalogRepo, slog.Default())
	dept := model.Department{Code: "IM", Name: "内科"}
	if e = db.Create(&dept).Error; e != nil {
		t.Fatal(e)
	}
	patient := model.Patient{RecordNo: "EMR001", Name: "王小明", Gender: "男", Age: 30, IDCard: "110101199001010001", Phone: "13800000001"}
	if e = db.Create(&patient).Error; e != nil {
		t.Fatal(e)
	}
	tmpl, e := catalogSvc.Template(dto.TemplateInput{Name: "门诊初诊模板", RecordType: "outpatient", Content: "主诉：\n现病史："})
	if e != nil {
		t.Fatal(e)
	}
	v1ID := tmpl.Versions[0].ID
	// 医生用第 1 版书写病历
	rec, e := recordSvc.Create(dto.RecordInput{PatientID: patient.ID, DepartmentID: dept.ID, RecordType: "outpatient", ChiefComplaint: "头痛三天", Diagnosis: "上呼吸道感染", TemplateVersionID: &v1ID}, 1)
	if e != nil {
		t.Fatal(e)
	}
	if rec.TemplateName != "门诊初诊模板" || rec.TemplateVersion != 1 || rec.TemplateContent != "主诉：\n现病史：" {
		t.Fatalf("record snapshot mismatch: %#v", rec)
	}
	// 管理员发布第 2 版
	if _, e = catalogSvc.PublishTemplateVersion(tmpl.ID, dto.TemplateVersionInput{Content: "主诉：\n现病史：\n既往史："}); e != nil {
		t.Fatal(e)
	}
	// 医生端只看到最新版
	views, e := catalogSvc.TemplateOptions()
	if e != nil {
		t.Fatal(e)
	}
	if len(views) != 1 || views[0].Version != 2 || views[0].Content != "主诉：\n现病史：\n既往史：" {
		t.Fatalf("doctor options must show latest version: %#v", views)
	}
	// 归档病历
	if _, e = recordSvc.Review(rec.ID, 1, true); e != nil {
		t.Fatal(e)
	}
	// 模板继续更新到第 3 版，已归档病历仍保留第 1 版快照
	if _, e = catalogSvc.PublishTemplateVersion(tmpl.ID, dto.TemplateVersionInput{Content: "主诉：\n现病史：\n既往史：\n过敏史："}); e != nil {
		t.Fatal(e)
	}
	archived, e := recordSvc.Get(rec.ID)
	if e != nil {
		t.Fatal(e)
	}
	if archived.Status != "archived" || archived.TemplateVersion != 1 || archived.TemplateContent != "主诉：\n现病史：" {
		t.Fatalf("archived record must keep original template snapshot: %#v", archived)
	}
	// 管理端可见全部历史版本
	full, e := catalogRepo.FindTemplateByID(tmpl.ID)
	if e != nil {
		t.Fatal(e)
	}
	if len(full.Versions) != 3 || full.Versions[0].Version != 3 || full.Versions[2].Version != 1 {
		t.Fatalf("template history incomplete: %#v", full.Versions)
	}
}
