package metadata

import "time"

type Field struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	DocTypeID    uint      `gorm:"not null;index" json:"doctype_id"`
	FieldName    string    `gorm:"size:140;not null" json:"field_name"`
	Label        string    `gorm:"size:140;not null" json:"label"`
	FieldType    string    `gorm:"size:64;not null" json:"field_type"`
	Options      string    `gorm:"type:text" json:"options"`
	Required     bool      `gorm:"default:false" json:"required"`
	Unique       bool      `gorm:"default:false" json:"unique"`
	ReadOnly     bool      `gorm:"default:false" json:"read_only"`
	Hidden       bool      `gorm:"default:false" json:"hidden"`
	InListView   bool      `gorm:"default:false" json:"in_list_view"`
	DefaultValue string    `gorm:"column:default_value;type:text" json:"default_value"`
	SortOrder    int       `gorm:"not null;default:0" json:"sort_order"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (Field) TableName() string {
	return "tab_docfields"
}

func SupportedFieldTypes() []string {
	return []string{"Data", "Text", "Int", "Float", "Currency", "Date", "Datetime", "Check", "Select", "Link", "Table"}
}
