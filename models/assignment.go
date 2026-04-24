package models

import (
	"time"

	"gorm.io/gorm"
)

type Assignment struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
	DocType    string         `gorm:"column:doc_type;size:140;not null;index:idx_assignment_document" json:"doctype"`
	DocName    string         `gorm:"column:doc_name;size:140;not null;index:idx_assignment_document" json:"docname"`
	AssignedTo string         `gorm:"size:140;not null;index" json:"assigned_to"`
	AssignedBy string         `gorm:"size:140;not null;default:'system'" json:"assigned_by"`
	Status     string         `gorm:"size:40;not null;default:'open';index" json:"status"`
	Note       string         `gorm:"type:text" json:"note"`
}

func (Assignment) TableName() string {
	return "tab_assignments"
}
