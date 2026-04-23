package models

import (
	"time"

	"gorm.io/gorm"
)

type File struct {
	ID                uint           `gorm:"primaryKey" json:"id"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
	Name              string         `gorm:"size:140;not null;uniqueIndex" json:"name"`
	OriginalName      string         `gorm:"size:255;not null" json:"original_name"`
	StoredName        string         `gorm:"size:255;not null" json:"stored_name"`
	StoragePath       string         `gorm:"size:500;not null" json:"storage_path"`
	FileURL           string         `gorm:"size:500;not null" json:"file_url"`
	Visibility        string         `gorm:"size:16;not null;default:'public'" json:"visibility"`
	ContentType       string         `gorm:"size:255" json:"content_type,omitempty"`
	Extension         string         `gorm:"size:32" json:"extension,omitempty"`
	SizeBytes         int64          `gorm:"not null;default:0" json:"size_bytes"`
	AttachedToDocType string         `gorm:"size:140" json:"attached_to_doctype,omitempty"`
	AttachedToName    string         `gorm:"size:140" json:"attached_to_name,omitempty"`
	AttachedToField   string         `gorm:"size:140" json:"attached_to_field,omitempty"`
	AltText           string         `gorm:"size:255" json:"alt_text,omitempty"`
	Attributes        string         `gorm:"type:jsonb" json:"attributes,omitempty"`
}

func (File) TableName() string {
	return "tab_files"
}
