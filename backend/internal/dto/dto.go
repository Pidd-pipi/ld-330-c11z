package dto

import "time"

type LoginRequest struct {
	Username string `json:"username" validate:"required,min=3,max=64"`
	Password string `json:"password" validate:"required,min=6,max=100"`
}
type PatientInput struct {
	Name           string `json:"name" validate:"required,max=64"`
	Gender         string `json:"gender" validate:"required,oneof=男 女 其他"`
	Age            int    `json:"age" validate:"gte=0,lte=150"`
	IDCard         string `json:"id_card" validate:"required,min=10,max=32"`
	Phone          string `json:"phone" validate:"required,min=6,max=32"`
	Allergies      string `json:"allergies"`
	MedicalHistory string `json:"medical_history"`
}
type RecordInput struct {
	PatientID      uint   `json:"patient_id" validate:"required"`
	DepartmentID   uint   `json:"department_id" validate:"required"`
	RecordType     string `json:"record_type" validate:"required,oneof=outpatient inpatient"`
	ChiefComplaint string `json:"chief_complaint" validate:"required"`
	PresentIllness string `json:"present_illness"`
	PastHistory    string `json:"past_history"`
	PhysicalExam   string `json:"physical_exam"`
	AuxiliaryExam  string `json:"auxiliary_exam"`
	Diagnosis      string `json:"diagnosis" validate:"required"`
	TreatmentPlan  string `json:"treatment_plan"`
	RichContent    string `json:"rich_content"`
	// TemplateVersionID 指向医生书写时选中的模板版本，保存时后端把该版本的
	// 名称、版本号、内容快照进病历，归档后不受模板后续更新影响。
	TemplateVersionID *uint `json:"template_version_id"`
}
type OrderInput struct {
	MedicalRecordID uint   `json:"medical_record_id" validate:"required"`
	Type            string `json:"type" validate:"required,oneof=long temporary"`
	Content         string `json:"content" validate:"required"`
}
type PrescriptionItemInput struct {
	DrugName      string `json:"drug_name" validate:"required"`
	Specification string `json:"specification"`
	Dosage        string `json:"dosage" validate:"required"`
	Frequency     string `json:"frequency" validate:"required"`
	Duration      string `json:"duration" validate:"required"`
}
type PrescriptionInput struct {
	MedicalRecordID uint                    `json:"medical_record_id" validate:"required"`
	Items           []PrescriptionItemInput `json:"items" validate:"required,min=1,dive"`
}
type DepartmentInput struct {
	Code        string `json:"code" validate:"required,max=32"`
	Name        string `json:"name" validate:"required,max=100"`
	Description string `json:"description"`
}
type UserInput struct {
	Username     string `json:"username" validate:"required,min=3"`
	Password     string `json:"password" validate:"required,min=6"`
	Name         string `json:"name" validate:"required"`
	Role         string `json:"role" validate:"required,oneof=admin doctor nurse"`
	DepartmentID *uint  `json:"department_id"`
}
type CatalogInput struct {
	Name          string `json:"name" validate:"required"`
	Specification string `json:"specification"`
	Unit          string `json:"unit"`
	Code          string `json:"code"`
	RecordType    string `json:"record_type"`
	Content       string `json:"content"`
}

// TemplateInput 用于新建模板，首个内容即第 1 版。
type TemplateInput struct {
	Name       string `json:"name" validate:"required,max=128"`
	RecordType string `json:"record_type" validate:"required,oneof=outpatient inpatient"`
	Content    string `json:"content" validate:"required"`
}

// TemplateVersionInput 用于在既有模板上发布新版本，历史版本保持不变。
type TemplateVersionInput struct {
	Content string `json:"content" validate:"required"`
}

// TemplateView 是医生书写病历时看到的模板选项，只暴露最新一版。
type TemplateView struct {
	ID         uint      `json:"id"`
	Name       string    `json:"name"`
	RecordType string    `json:"record_type"`
	VersionID  uint      `json:"version_id"`
	Version    int       `json:"version"`
	Content    string    `json:"content"`
	UpdatedAt  time.Time `json:"updated_at"`
}
