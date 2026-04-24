package metadata

import "time"

type DocType struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Name          string    `gorm:"size:140;not null;uniqueIndex" json:"name"`
	Module        string    `gorm:"size:140;not null;default:Core" json:"module"`
	StorageTable  string    `gorm:"column:table_name;size:140;not null;uniqueIndex" json:"table_name"`
	IsSubmittable bool      `gorm:"default:false" json:"is_submittable"`
	IsChildTable  bool      `gorm:"column:is_child_table;default:false" json:"is_child_table"`
	Fields        []Field   `gorm:"foreignKey:DocTypeID;constraint:OnDelete:CASCADE" json:"fields"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (DocType) TableName() string {
	return "tab_doctypes"
}
