package models

import "time"

type AuditLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	DocType   string    `gorm:"column:doc_type;size:140;not null;index:idx_audit_document" json:"doctype"`
	DocName   string    `gorm:"column:doc_name;size:140;not null;index:idx_audit_document" json:"docname"`
	Action    string    `gorm:"size:40;not null;index" json:"action"`
	Actor     string    `gorm:"size:140;not null;default:'system'" json:"actor"`
	Summary   string    `gorm:"type:text" json:"summary"`
	Metadata  string    `gorm:"type:jsonb;not null;default:'{}'" json:"metadata"`
}

func (AuditLog) TableName() string {
	return "tab_audit_logs"
}
