package model

import "time"

type Drug struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Name          string    `gorm:"uniqueIndex;size:128;not null" json:"name"`
	Specification string    `gorm:"size:128" json:"specification"`
	Unit          string    `gorm:"size:32" json:"unit"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
type DiagnosisCode struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Code      string    `gorm:"uniqueIndex;size:32;not null" json:"code"`
	Name      string    `gorm:"size:128;not null" json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RecordTemplate 是模板主体，内容按版本存放在 RecordTemplateVersion 中，
// 每次修改只新增版本行，历史版本不可变，便于病历追溯当时使用的模板文字。
type RecordTemplate struct {
	ID         uint                    `gorm:"primaryKey" json:"id"`
	Name       string                  `gorm:"uniqueIndex;size:128;not null" json:"name"`
	RecordType string                  `gorm:"size:20;not null" json:"record_type"`
	Versions   []RecordTemplateVersion `gorm:"foreignKey:TemplateID;constraint:OnDelete:CASCADE" json:"versions"`
	CreatedAt  time.Time               `json:"created_at"`
	UpdatedAt  time.Time               `json:"updated_at"`
}
type RecordTemplateVersion struct {
	ID         uint            `gorm:"primaryKey" json:"id"`
	TemplateID uint            `gorm:"uniqueIndex:idx_template_version;not null" json:"template_id"`
	Template   *RecordTemplate `gorm:"foreignKey:TemplateID" json:"template,omitempty"`
	Version    int             `gorm:"uniqueIndex:idx_template_version;not null" json:"version"`
	Content    string          `gorm:"type:text" json:"content"`
	CreatedAt  time.Time       `json:"created_at"`
}
type AuditLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index" json:"user_id"`
	Username  string    `gorm:"size:64" json:"username"`
	Action    string    `gorm:"size:100;not null" json:"action"`
	Resource  string    `gorm:"size:100" json:"resource"`
	Detail    string    `gorm:"type:text" json:"detail"`
	CreatedAt time.Time `json:"created_at"`
}
