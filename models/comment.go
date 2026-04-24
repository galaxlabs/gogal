package models

import (
	"time"

	"gorm.io/gorm"
)

type Comment struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	DocType   string         `gorm:"column:doc_type;size:140;not null;index:idx_comment_document" json:"doctype"`
	DocName   string         `gorm:"column:doc_name;size:140;not null;index:idx_comment_document" json:"docname"`
	Author    string         `gorm:"size:140;not null;default:'system'" json:"author"`
	Body      string         `gorm:"type:text;not null" json:"body"`
}

func (Comment) TableName() string {
	return "tab_comments"
}
