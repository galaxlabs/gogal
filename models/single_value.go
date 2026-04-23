package models

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SingleValue struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DocType   string    `gorm:"column:doctype;size:140;not null;uniqueIndex:idx_single_value_key" json:"doctype"`
	Field     string    `gorm:"column:field;size:140;not null;uniqueIndex:idx_single_value_key" json:"field"`
	Value     string    `gorm:"column:value;type:text" json:"value"`
}

func (SingleValue) TableName() string {
	return "tab_singles"
}

func LoadSingleDocument(db *gorm.DB, docType *DocType) (map[string]any, bool, error) {
	fieldMap := make(map[string]DocField, len(docType.Fields))
	document := make(map[string]any, len(docType.Fields)+3)
	document["name"] = docType.Name

	for _, field := range docType.Fields {
		if !IsStoredInParentTable(field) {
			continue
		}
		fieldMap[field.FieldName] = field
		if field.DefaultValue != "" {
			coercedDefault, err := CoerceFieldValue(field, field.DefaultValue)
			if err != nil {
				return nil, false, fmt.Errorf("field %q default value is invalid: %w", field.FieldName, err)
			}
			document[field.FieldName] = coercedDefault
			continue
		}

		if field.FieldType == "Check" {
			document[field.FieldName] = false
		}
	}

	rows := make([]SingleValue, 0, len(docType.Fields))
	if err := db.Where("doctype = ?", docType.Name).Order("field ASC").Find(&rows).Error; err != nil {
		return nil, false, err
	}

	if len(rows) == 0 {
		return document, false, nil
	}

	createdAt := rows[0].CreatedAt
	updatedAt := rows[0].UpdatedAt
	for _, row := range rows {
		field, ok := fieldMap[row.Field]
		if !ok {
			continue
		}

		parsedValue, err := DeserializeSingleFieldValue(field, row.Value)
		if err != nil {
			return nil, true, fmt.Errorf("field %q: %w", row.Field, err)
		}
		document[row.Field] = parsedValue

		if row.CreatedAt.Before(createdAt) {
			createdAt = row.CreatedAt
		}
		if row.UpdatedAt.After(updatedAt) {
			updatedAt = row.UpdatedAt
		}
	}

	document["created_at"] = createdAt
	document["updated_at"] = updatedAt
	return document, true, nil
}

func SaveSingleDocument(db *gorm.DB, docType *DocType, values map[string]any) (bool, error) {
	var existingCount int64
	if err := db.Model(&SingleValue{}).Where("doctype = ?", docType.Name).Count(&existingCount).Error; err != nil {
		return false, err
	}

	for _, field := range docType.Fields {
		if !IsStoredInParentTable(field) {
			continue
		}

		serializedValue, err := SerializeSingleFieldValue(field, values[field.FieldName])
		if err != nil {
			return existingCount > 0, fmt.Errorf("field %q: %w", field.FieldName, err)
		}

		if serializedValue == "" {
			if err := db.Where("doctype = ? AND field = ?", docType.Name, field.FieldName).Delete(&SingleValue{}).Error; err != nil {
				return existingCount > 0, err
			}
			continue
		}

		record := SingleValue{
			DocType: docType.Name,
			Field:   field.FieldName,
			Value:   serializedValue,
		}

		if err := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "doctype"}, {Name: "field"}},
			DoUpdates: clause.Assignments(map[string]any{"value": serializedValue, "updated_at": gorm.Expr("NOW()")}),
		}).Create(&record).Error; err != nil {
			return existingCount > 0, err
		}
	}

	return existingCount > 0, nil
}

func DeleteSingleDocument(db *gorm.DB, docTypeName string) error {
	return db.Where("doctype = ?", strings.TrimSpace(docTypeName)).Delete(&SingleValue{}).Error
}

func SerializeSingleFieldValue(field DocField, value any) (string, error) {
	coercedValue, err := CoerceFieldValue(field, value)
	if err != nil {
		return "", err
	}

	if isEmptyValue(coercedValue) {
		return "", nil
	}

	switch typed := coercedValue.(type) {
	case nil:
		return "", nil
	case string:
		return typed, nil
	case bool:
		if typed {
			return "1", nil
		}
		return "0", nil
	case int64:
		return strconv.FormatInt(typed, 10), nil
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64), nil
	case time.Time:
		return typed.UTC().Format(time.RFC3339), nil
	default:
		return fmt.Sprint(typed), nil
	}
}

func DeserializeSingleFieldValue(field DocField, value string) (any, error) {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		if field.FieldType == "Check" {
			return false, nil
		}
		return "", nil
	}

	return CoerceFieldValue(field, trimmedValue)
}
