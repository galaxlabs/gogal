package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type PreparedChildTable struct {
	Field        DocField
	ChildDocType *DocType
	Rows         []map[string]any
}

type PreparedDocumentMutation struct {
	Values      map[string]any
	ChildTables []PreparedChildTable
}

func PrepareDocumentMutation(db *gorm.DB, docType *DocType, payload map[string]any, isCreate bool) (*PreparedDocumentMutation, error) {
	fieldMap := make(map[string]DocField, len(docType.Fields))
	for _, field := range docType.Fields {
		fieldMap[field.FieldName] = field
	}

	preparedValues := make(map[string]any, len(payload))
	childTables := make([]PreparedChildTable, 0)

	for rawFieldName, rawValue := range payload {
		fieldName := NormalizeIdentifier(rawFieldName)
		if fieldName == "" {
			return nil, fmt.Errorf("invalid field name %q", rawFieldName)
		}

		field, ok := fieldMap[fieldName]
		if !ok {
			return nil, fmt.Errorf("field %q does not exist in DocType %q", rawFieldName, docType.Name)
		}

		if field.ReadOnly {
			return nil, fmt.Errorf("field %q is read only", field.FieldName)
		}

		if field.FieldType == "Table" {
			childTable, err := prepareChildTable(db, field, rawValue)
			if err != nil {
				return nil, fmt.Errorf("field %q: %w", field.FieldName, err)
			}
			if field.Required && len(childTable.Rows) == 0 {
				return nil, fmt.Errorf("field %q is required", field.FieldName)
			}
			childTables = append(childTables, childTable)
			continue
		}

		coercedValue, err := CoerceFieldValue(field, rawValue)
		if err != nil {
			return nil, fmt.Errorf("field %q: %w", field.FieldName, err)
		}

		if field.FieldType == "Link" && !isEmptyValue(coercedValue) {
			if err := ValidateLinkValue(db, field, coercedValue); err != nil {
				return nil, fmt.Errorf("field %q: %w", field.FieldName, err)
			}
		}

		if field.Required && isEmptyValue(coercedValue) {
			return nil, fmt.Errorf("field %q is required", field.FieldName)
		}

		preparedValues[field.FieldName] = coercedValue
	}

	if isCreate {
		for _, field := range docType.Fields {
			if field.FieldType == "Table" {
				if field.Required {
					found := false
					for _, childTable := range childTables {
						if childTable.Field.FieldName == field.FieldName {
							found = true
							break
						}
					}
					if !found {
						return nil, fmt.Errorf("field %q is required", field.FieldName)
					}
				}
				continue
			}

			if _, exists := preparedValues[field.FieldName]; exists {
				continue
			}

			if field.DefaultValue != "" {
				coercedDefault, err := CoerceFieldValue(field, field.DefaultValue)
				if err != nil {
					return nil, fmt.Errorf("field %q default value is invalid: %w", field.FieldName, err)
				}
				if field.FieldType == "Link" && !isEmptyValue(coercedDefault) {
					if err := ValidateLinkValue(db, field, coercedDefault); err != nil {
						return nil, fmt.Errorf("field %q default value is invalid: %w", field.FieldName, err)
					}
				}
				preparedValues[field.FieldName] = coercedDefault
			}

			if field.Required && isEmptyValue(preparedValues[field.FieldName]) {
				return nil, fmt.Errorf("field %q is required", field.FieldName)
			}
		}
	}

	return &PreparedDocumentMutation{
		Values:      preparedValues,
		ChildTables: childTables,
	}, nil
}

func ValidateLinkValue(db *gorm.DB, field DocField, value any) error {
	linkedName := strings.TrimSpace(fmt.Sprint(value))
	if linkedName == "" {
		return nil
	}

	targetDocTypeName := ResolveTargetDocTypeName(field)
	if targetDocTypeName == "" {
		return fmt.Errorf("options must specify a target DocType")
	}

	targetDocType, err := LoadDocTypeByName(db, targetDocTypeName, false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("target DocType %q does not exist", targetDocTypeName)
		}
		return fmt.Errorf("load target DocType %q: %w", targetDocTypeName, err)
	}

	var count int64
	if err := db.Table(targetDocType.StorageTable).
		Where("name = ? AND deleted_at IS NULL", linkedName).
		Count(&count).Error; err != nil {
		return fmt.Errorf("validate linked record %q: %w", linkedName, err)
	}

	if count == 0 {
		return fmt.Errorf("linked record %q was not found in %q", linkedName, targetDocType.Name)
	}

	return nil
}

func prepareChildTable(db *gorm.DB, field DocField, rawValue any) (PreparedChildTable, error) {
	targetDocTypeName := ResolveTargetDocTypeName(field)
	if targetDocTypeName == "" {
		return PreparedChildTable{}, fmt.Errorf("options must specify a child DocType")
	}

	childDocType, err := LoadDocTypeByName(db, targetDocTypeName, true)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return PreparedChildTable{}, fmt.Errorf("child DocType %q does not exist", targetDocTypeName)
		}
		return PreparedChildTable{}, fmt.Errorf("load child DocType %q: %w", targetDocTypeName, err)
	}

	if !childDocType.IsChildTable {
		return PreparedChildTable{}, fmt.Errorf("DocType %q must be marked is_child_table=true before it can be used in a Table field", childDocType.Name)
	}

	rawRows, err := parseChildTableRows(rawValue)
	if err != nil {
		return PreparedChildTable{}, err
	}

	preparedRows := make([]map[string]any, 0, len(rawRows))
	for index, row := range rawRows {
		rowPayload := sanitizeChildRowPayload(row)
		rowName, err := extractChildRowName(rowPayload, childDocType.Name)
		if err != nil {
			return PreparedChildTable{}, fmt.Errorf("row %d: %w", index+1, err)
		}

		mutation, err := PrepareDocumentMutation(db, childDocType, rowPayload, true)
		if err != nil {
			return PreparedChildTable{}, fmt.Errorf("row %d: %w", index+1, err)
		}

		if len(mutation.ChildTables) > 0 {
			return PreparedChildTable{}, fmt.Errorf("row %d: nested Table fields are not supported inside child table rows", index+1)
		}

		preparedRow := make(map[string]any, len(mutation.Values)+1)
		preparedRow["name"] = rowName
		for key, value := range mutation.Values {
			preparedRow[key] = value
		}
		preparedRows = append(preparedRows, preparedRow)
	}

	return PreparedChildTable{
		Field:        field,
		ChildDocType: childDocType,
		Rows:         preparedRows,
	}, nil
}

func parseChildTableRows(value any) ([]map[string]any, error) {
	if value == nil {
		return []map[string]any{}, nil
	}

	switch typed := value.(type) {
	case string:
		raw := strings.TrimSpace(typed)
		if raw == "" {
			return []map[string]any{}, nil
		}
		var rows []map[string]any
		if err := json.Unmarshal([]byte(raw), &rows); err != nil {
			return nil, fmt.Errorf("must be an array of child row objects")
		}
		return rows, nil
	case []map[string]any:
		return typed, nil
	case []any:
		rows := make([]map[string]any, 0, len(typed))
		for _, item := range typed {
			row, ok := item.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("must be an array of child row objects")
			}
			rows = append(rows, row)
		}
		return rows, nil
	default:
		bytes, err := json.Marshal(typed)
		if err != nil {
			return nil, fmt.Errorf("must be an array of child row objects")
		}
		var rows []map[string]any
		if err := json.Unmarshal(bytes, &rows); err != nil {
			return nil, fmt.Errorf("must be an array of child row objects")
		}
		return rows, nil
	}
}

func sanitizeChildRowPayload(row map[string]any) map[string]any {
	cleaned := make(map[string]any, len(row))
	for key, value := range row {
		normalized := NormalizeIdentifier(key)
		switch normalized {
		case "", "id", "created_at", "updated_at", "deleted_at", "parent", "parenttype", "parentfield", "idx":
			continue
		default:
			cleaned[normalized] = value
		}
	}
	return cleaned
}

func extractChildRowName(row map[string]any, docTypeName string) (string, error) {
	if rawName, hasName := row["name"]; hasName {
		delete(row, "name")
		if rawName == nil {
			return "", fmt.Errorf("document name is required")
		}
		return NormalizeDocumentName(fmt.Sprint(rawName))
	}

	return GenerateDocumentName(docTypeName)
}
