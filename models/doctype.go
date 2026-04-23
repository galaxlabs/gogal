package models

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
)

var identifierPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

var reservedColumnNames = map[string]struct{}{
	"id":         {},
	"name":       {},
	"created_at": {},
	"updated_at": {},
	"deleted_at": {},
}

var supportedFieldTypes = map[string]string{
	"Check":       "BOOLEAN",
	"Currency":    "NUMERIC(18,6)",
	"Data":        "TEXT",
	"Date":        "DATE",
	"Datetime":    "TIMESTAMPTZ",
	"DynamicLink": "TEXT",
	"Float":       "NUMERIC(18,6)",
	"Int":         "INTEGER",
	"JSON":        "JSONB",
	"Link":        "TEXT",
	"Long Text":   "TEXT",
	"Percent":     "NUMERIC(8,4)",
	"Select":      "TEXT",
	"Small Text":  "TEXT",
	"Text":        "TEXT",
	"Time":        "TIME",
}

type CreateDocTypeRequest struct {
	Name         string     `json:"doctype" binding:"required"`
	Label        string     `json:"label"`
	Module       string     `json:"module"`
	TableName    string     `json:"table_name"`
	Description  string     `json:"description"`
	IsSingle     bool       `json:"is_single"`
	TrackChanges *bool      `json:"track_changes"`
	Fields       []DocField `json:"fields" binding:"required"`
}

type DocType struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	Name         string         `gorm:"size:140;not null;uniqueIndex" json:"doctype"`
	Label        string         `gorm:"size:140;not null" json:"label"`
	Module       string         `gorm:"size:140;not null;default:'Core'" json:"module"`
	StorageTable string         `gorm:"column:table_name;size:140;not null;uniqueIndex" json:"table_name"`
	Description  string         `gorm:"type:text" json:"description,omitempty"`
	IsSingle     bool           `gorm:"default:false" json:"is_single"`
	IsSystem     bool           `gorm:"default:false" json:"is_system"`
	TrackChanges bool           `gorm:"default:true" json:"track_changes"`
	Fields       []DocField     `gorm:"foreignKey:DocTypeID;constraint:OnDelete:CASCADE" json:"fields"`
}

type DocField struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	DocTypeID    uint           `gorm:"not null;uniqueIndex:idx_docfield_doctype_field" json:"doctype_id"`
	FieldName    string         `gorm:"size:140;not null;uniqueIndex:idx_docfield_doctype_field" json:"fieldname"`
	Label        string         `gorm:"size:140;not null" json:"label"`
	FieldType    string         `gorm:"size:64;not null" json:"fieldtype"`
	Options      string         `gorm:"type:text" json:"options,omitempty"`
	DefaultValue string         `gorm:"column:default_value;type:text" json:"default,omitempty"`
	Required     bool           `gorm:"default:false" json:"reqd"`
	ReadOnly     bool           `gorm:"default:false" json:"read_only"`
	Hidden       bool           `gorm:"default:false" json:"hidden"`
	Unique       bool           `gorm:"default:false" json:"unique"`
	InListView   bool           `gorm:"default:false" json:"in_list_view"`
	SortOrder    int            `gorm:"not null;default:0;index" json:"sort_order"`
}

func (DocType) TableName() string {
	return "tab_doctypes"
}

func (DocField) TableName() string {
	return "tab_docfields"
}

func SupportedFieldTypes() []string {
	fieldTypes := make([]string, 0, len(supportedFieldTypes))
	for fieldType := range supportedFieldTypes {
		fieldTypes = append(fieldTypes, fieldType)
	}
	sort.Strings(fieldTypes)
	return fieldTypes
}

func NewDocTypeFromRequest(req CreateDocTypeRequest) (*DocType, error) {
	trackChanges := true
	if req.TrackChanges != nil {
		trackChanges = *req.TrackChanges
	}

	docType := &DocType{
		Name:         strings.TrimSpace(req.Name),
		Label:        strings.TrimSpace(req.Label),
		Module:       strings.TrimSpace(req.Module),
		StorageTable: strings.TrimSpace(req.TableName),
		Description:  strings.TrimSpace(req.Description),
		IsSingle:     req.IsSingle,
		TrackChanges: trackChanges,
		Fields:       req.Fields,
	}

	if err := docType.Normalize(); err != nil {
		return nil, err
	}

	return docType, nil
}

func (d *DocType) Normalize() error {
	d.Name = strings.TrimSpace(d.Name)
	d.Label = strings.TrimSpace(d.Label)
	d.Module = strings.TrimSpace(d.Module)
	d.StorageTable = strings.TrimSpace(d.StorageTable)
	d.Description = strings.TrimSpace(d.Description)

	if d.Name == "" {
		return fmt.Errorf("doctype name is required")
	}

	if d.Label == "" {
		d.Label = d.Name
	}

	if d.Module == "" {
		d.Module = "Core"
	}

	if len(d.Fields) == 0 {
		return fmt.Errorf("at least one field is required")
	}

	if d.StorageTable == "" {
		d.StorageTable = BuildStorageTableName(d.Name)
	} else {
		d.StorageTable = NormalizeIdentifier(d.StorageTable)
	}

	seenFields := make(map[string]struct{}, len(d.Fields))
	for index := range d.Fields {
		field := &d.Fields[index]
		field.SortOrder = index + 1

		if err := field.Normalize(); err != nil {
			return fmt.Errorf("invalid field at position %d: %w", index+1, err)
		}

		if _, exists := seenFields[field.FieldName]; exists {
			return fmt.Errorf("duplicate fieldname %q", field.FieldName)
		}

		if !d.IsSystem {
			if _, reserved := reservedColumnNames[field.FieldName]; reserved {
				return fmt.Errorf("fieldname %q is reserved by the framework", field.FieldName)
			}
		}

		seenFields[field.FieldName] = struct{}{}
	}

	if err := ValidateIdentifier(d.StorageTable); err != nil {
		return fmt.Errorf("invalid table_name: %w", err)
	}

	return nil
}

func (f *DocField) Normalize() error {
	f.FieldName = NormalizeIdentifier(f.FieldName)
	f.Label = strings.TrimSpace(f.Label)
	f.FieldType = strings.TrimSpace(f.FieldType)
	f.Options = strings.TrimSpace(f.Options)
	f.DefaultValue = strings.TrimSpace(f.DefaultValue)

	if f.FieldName == "" {
		return fmt.Errorf("fieldname is required")
	}

	if f.Label == "" {
		f.Label = HumanizeIdentifier(f.FieldName)
	}

	if err := ValidateIdentifier(f.FieldName); err != nil {
		return fmt.Errorf("invalid fieldname %q: %w", f.FieldName, err)
	}

	if _, ok := supportedFieldTypes[f.FieldType]; !ok {
		return fmt.Errorf("unsupported fieldtype %q", f.FieldType)
	}

	return nil
}

func BuildStorageTableName(doctypeName string) string {
	identifier := NormalizeIdentifier(doctypeName)
	if identifier == "" {
		return ""
	}

	return fmt.Sprintf("tab_%s", identifier)
}

func NormalizeIdentifier(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.ReplaceAll(value, "-", "_")
	value = strings.ReplaceAll(value, " ", "_")

	var builder strings.Builder
	builder.Grow(len(value))
	for _, char := range value {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '_' {
			builder.WriteRune(char)
		}
	}

	return strings.Trim(builder.String(), "_")
}

func ValidateIdentifier(value string) error {
	if !identifierPattern.MatchString(value) {
		return fmt.Errorf("must start with a letter and contain only lowercase letters, numbers, and underscores")
	}

	return nil
}

func HumanizeIdentifier(identifier string) string {
	parts := strings.Fields(strings.ReplaceAll(identifier, "_", " "))
	for index, part := range parts {
		if part == "" {
			continue
		}
		parts[index] = strings.ToUpper(part[:1]) + part[1:]
	}

	return strings.Join(parts, " ")
}

func FieldDatabaseType(fieldType string) (string, bool) {
	databaseType, ok := supportedFieldTypes[fieldType]
	return databaseType, ok
}

func SystemDocTypes() []DocType {
	return []DocType{
		{
			Name:         "DocType",
			Label:        "DocType",
			Module:       "Core",
			StorageTable: "tab_doctypes",
			Description:  "System metadata that defines business document types.",
			IsSystem:     true,
			TrackChanges: true,
			Fields: []DocField{
				{FieldName: "name", Label: "Name", FieldType: "Data", Required: true, Unique: true},
				{FieldName: "label", Label: "Label", FieldType: "Data", Required: true},
				{FieldName: "module", Label: "Module", FieldType: "Data", Required: true},
				{FieldName: "table_name", Label: "Storage Table", FieldType: "Data", Required: true, Unique: true},
				{FieldName: "description", Label: "Description", FieldType: "Text"},
				{FieldName: "is_single", Label: "Is Single", FieldType: "Check"},
				{FieldName: "is_system", Label: "Is System", FieldType: "Check"},
				{FieldName: "track_changes", Label: "Track Changes", FieldType: "Check"},
			},
		},
		{
			Name:         "DocField",
			Label:        "DocField",
			Module:       "Core",
			StorageTable: "tab_docfields",
			Description:  "Child metadata rows that define fields for a DocType.",
			IsSystem:     true,
			TrackChanges: true,
			Fields: []DocField{
				{FieldName: "fieldname", Label: "Field Name", FieldType: "Data", Required: true},
				{FieldName: "label", Label: "Label", FieldType: "Data", Required: true},
				{FieldName: "fieldtype", Label: "Field Type", FieldType: "Select", Required: true},
				{FieldName: "options", Label: "Options", FieldType: "Text"},
				{FieldName: "default_value", Label: "Default Value", FieldType: "Data"},
				{FieldName: "reqd", Label: "Required", FieldType: "Check"},
				{FieldName: "read_only", Label: "Read Only", FieldType: "Check"},
				{FieldName: "hidden", Label: "Hidden", FieldType: "Check"},
				{FieldName: "unique", Label: "Unique", FieldType: "Check"},
				{FieldName: "in_list_view", Label: "In List View", FieldType: "Check"},
				{FieldName: "sort_order", Label: "Sort Order", FieldType: "Int"},
			},
		},
	}
}
