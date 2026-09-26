package main

import (
	"github.com/blueship581/gbemr/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"testing"
)

// TestMigrateLegacyTemplateContent 模拟旧版数据库（record_templates 带 content 列、无版本表），
// 验证迁移后旧模板内容成为第 1 版，模板继续可选。
func TestMigrateLegacyTemplateContent(t *testing.T) {
	db, e := gorm.Open(sqlite.Open("file:legacymigrate?mode=memory&cache=shared"), &gorm.Config{})
	if e != nil {
		t.Fatal(e)
	}
	if e = db.Exec(`CREATE TABLE record_templates (id integer primary key autoincrement, name text, record_type text, content text, created_at datetime, updated_at datetime)`).Error; e != nil {
		t.Fatal(e)
	}
	if e = db.Exec(`INSERT INTO record_templates (name, record_type, content, created_at, updated_at) VALUES ('旧门诊模板', 'outpatient', '旧内容：主诉', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`).Error; e != nil {
		t.Fatal(e)
	}
	if e = migrateAndSeed(db); e != nil {
		t.Fatal(e)
	}
	if db.Migrator().HasColumn(&model.RecordTemplate{}, "content") {
		t.Fatal("legacy content column should be dropped")
	}
	var tmpl model.RecordTemplate
	if e = db.Preload("Versions").Where("name = ?", "旧门诊模板").First(&tmpl).Error; e != nil {
		t.Fatal(e)
	}
	if len(tmpl.Versions) != 1 || tmpl.Versions[0].Version != 1 || tmpl.Versions[0].Content != "旧内容：主诉" {
		t.Fatalf("legacy content must become version 1: %#v", tmpl.Versions)
	}
	// 迁移是幂等的，再次执行不会重复插入版本
	if e = migrateAndSeed(db); e != nil {
		t.Fatal(e)
	}
	var count int64
	db.Model(&model.RecordTemplateVersion{}).Where("template_id = ?", tmpl.ID).Count(&count)
	if count != 1 {
		t.Fatalf("expected 1 version after re-migrate, got %d", count)
	}
}
