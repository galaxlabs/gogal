package models

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

const documentNameMaxLength = 140

var systemQueryFields = map[string]DocField{
	"id":         {FieldName: "id", FieldType: "Int", Label: "ID"},
	"name":       {FieldName: "name", FieldType: "Data", Label: "Name"},
	"created_at": {FieldName: "created_at", FieldType: "Datetime", Label: "Created At"},
	"updated_at": {FieldName: "updated_at", FieldType: "Datetime", Label: "Updated At"},
}

func PrepareDocumentPayload(docType *DocType, payload map[string]any, isCreate bool) (map[string]any, error) {
	fieldMap := make(map[string]DocField, len(docType.Fields))
	for _, field := range docType.Fields {
		fieldMap[field.FieldName] = field
	}

	prepared := make(map[string]any, len(payload))
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

		coercedValue, err := CoerceFieldValue(field, rawValue)
		if err != nil {
			return nil, fmt.Errorf("field %q: %w", field.FieldName, err)
		}

		if field.Required && isEmptyValue(coercedValue) {
			return nil, fmt.Errorf("field %q is required", field.FieldName)
		}

		prepared[field.FieldName] = coercedValue
	}

	if isCreate {
		for _, field := range docType.Fields {
			if _, exists := prepared[field.FieldName]; exists {
				continue
			}

			if field.DefaultValue != "" {
				coercedDefault, err := CoerceFieldValue(field, field.DefaultValue)
				if err != nil {
					return nil, fmt.Errorf("field %q default value is invalid: %w", field.FieldName, err)
				}
				prepared[field.FieldName] = coercedDefault
			}

			if field.Required && isEmptyValue(prepared[field.FieldName]) {
				return nil, fmt.Errorf("field %q is required", field.FieldName)
			}
		}
	}

	return prepared, nil
}

func CoerceFieldValue(field DocField, value any) (any, error) {
	if value == nil {
		return nil, nil
	}

	switch field.FieldType {
	case "Attach", "Attach Image", "Data", "DynamicLink", "Image", "Link", "Long Text", "Select", "Small Text", "Text":
		return strings.TrimSpace(fmt.Sprint(value)), nil
	case "Check":
		return coerceBool(value)
	case "Int":
		return coerceInt(value)
	case "Float", "Currency", "Percent":
		return coerceFloat(value)
	case "Date":
		return coerceDate(value)
	case "Datetime":
		return coerceDateTime(value)
	case "Time":
		return coerceTime(value)
	case "JSON":
		return coerceJSON(value)
	default:
		return nil, fmt.Errorf("unsupported fieldtype %q", field.FieldType)
	}
}

func GenerateDocumentName(docTypeName string) (string, error) {
	base := NormalizeIdentifier(docTypeName)
	if base == "" {
		base = "doc"
	}

	randomBytes := make([]byte, 6)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("generate random suffix: %w", err)
	}

	name := fmt.Sprintf("%s-%s", base, hex.EncodeToString(randomBytes))
	if len(name) > documentNameMaxLength {
		name = name[:documentNameMaxLength]
	}

	return name, nil
}

func NormalizeDocumentName(value string) (string, error) {
	name := strings.TrimSpace(value)
	if name == "" {
		return "", fmt.Errorf("document name is required")
	}

	if len(name) > documentNameMaxLength {
		return "", fmt.Errorf("document name cannot exceed %d characters", documentNameMaxLength)
	}

	return name, nil
}

func DocumentSelectColumns(docType *DocType) []string {
	columns := make([]string, 0, len(docType.Fields)+4)
	columns = append(columns, "id", "name", "created_at", "updated_at")
	for _, field := range docType.Fields {
		if !IsStoredInParentTable(field) {
			continue
		}
		columns = append(columns, field.FieldName)
	}

	return columns
}

func QueryableField(docType *DocType, fieldName string) (DocField, bool) {
	normalizedFieldName := NormalizeIdentifier(fieldName)
	if field, ok := systemQueryFields[normalizedFieldName]; ok {
		return field, true
	}

	for _, field := range docType.Fields {
		if field.FieldName == normalizedFieldName {
			return field, true
		}
	}

	return DocField{}, false
}

func SearchableColumns(docType *DocType) []string {
	columns := []string{"name"}
	for _, field := range docType.Fields {
		if !IsStoredInParentTable(field) {
			continue
		}
		if IsTextLikeFieldType(field.FieldType) {
			columns = append(columns, field.FieldName)
		}
	}

	return columns
}

func IsTextLikeFieldType(fieldType string) bool {
	switch fieldType {
	case "Attach", "Attach Image", "Data", "DynamicLink", "Image", "Link", "Long Text", "Select", "Small Text", "Text":
		return true
	default:
		return false
	}
}

func IsSortableField(field DocField) bool {
	return field.FieldType != "JSON" && field.FieldType != "Table"
}

func isEmptyValue(value any) bool {
	if value == nil {
		return true
	}

	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed) == ""
	}

	return false
}

func coerceBool(value any) (bool, error) {
	switch typed := value.(type) {
	case bool:
		return typed, nil
	case string:
		normalized := strings.TrimSpace(strings.ToLower(typed))
		switch normalized {
		case "1", "true", "yes", "y", "on":
			return true, nil
		case "0", "false", "no", "n", "off", "":
			return false, nil
		}
	case float64:
		return typed != 0, nil
	case float32:
		return typed != 0, nil
	case int:
		return typed != 0, nil
	case int64:
		return typed != 0, nil
	case json.Number:
		floatValue, err := typed.Float64()
		if err != nil {
			return false, fmt.Errorf("must be a boolean")
		}
		return floatValue != 0, nil
	}

	return false, fmt.Errorf("must be a boolean")
}

func coerceInt(value any) (int64, error) {
	switch typed := value.(type) {
	case int:
		return int64(typed), nil
	case int8:
		return int64(typed), nil
	case int16:
		return int64(typed), nil
	case int32:
		return int64(typed), nil
	case int64:
		return typed, nil
	case float64:
		if math.Trunc(typed) != typed {
			return 0, fmt.Errorf("must be a whole number")
		}
		return int64(typed), nil
	case json.Number:
		parsed, err := typed.Int64()
		if err == nil {
			return parsed, nil
		}
		floatValue, floatErr := typed.Float64()
		if floatErr != nil || math.Trunc(floatValue) != floatValue {
			return 0, fmt.Errorf("must be a whole number")
		}
		return int64(floatValue), nil
	case string:
		parsed, err := strconv.ParseInt(strings.TrimSpace(typed), 10, 64)
		if err != nil {
			return 0, fmt.Errorf("must be a whole number")
		}
		return parsed, nil
	}

	return 0, fmt.Errorf("must be a whole number")
}

func coerceFloat(value any) (float64, error) {
	switch typed := value.(type) {
	case float64:
		return typed, nil
	case float32:
		return float64(typed), nil
	case int:
		return float64(typed), nil
	case int64:
		return float64(typed), nil
	case json.Number:
		parsed, err := typed.Float64()
		if err != nil {
			return 0, fmt.Errorf("must be numeric")
		}
		return parsed, nil
	case string:
		parsed, err := strconv.ParseFloat(strings.TrimSpace(typed), 64)
		if err != nil {
			return 0, fmt.Errorf("must be numeric")
		}
		return parsed, nil
	}

	return 0, fmt.Errorf("must be numeric")
}

func coerceDate(value any) (string, error) {
	switch typed := value.(type) {
	case time.Time:
		return typed.UTC().Format("2006-01-02"), nil
	case string:
		parsed, err := time.Parse("2006-01-02", strings.TrimSpace(typed))
		if err != nil {
			return "", fmt.Errorf("must use YYYY-MM-DD format")
		}
		return parsed.Format("2006-01-02"), nil
	}

	return "", fmt.Errorf("must use YYYY-MM-DD format")
}

func coerceDateTime(value any) (time.Time, error) {
	switch typed := value.(type) {
	case time.Time:
		return typed.UTC(), nil
	case string:
		raw := strings.TrimSpace(typed)
		layouts := []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02T15:04:05"}
		for _, layout := range layouts {
			parsed, err := time.Parse(layout, raw)
			if err == nil {
				return parsed.UTC(), nil
			}
		}
	}

	return time.Time{}, fmt.Errorf("must use RFC3339 or YYYY-MM-DD HH:MM:SS format")
}

func coerceTime(value any) (string, error) {
	switch typed := value.(type) {
	case time.Time:
		return typed.Format("15:04:05"), nil
	case string:
		raw := strings.TrimSpace(typed)
		layouts := []string{"15:04:05", "15:04"}
		for _, layout := range layouts {
			parsed, err := time.Parse(layout, raw)
			if err == nil {
				return parsed.Format("15:04:05"), nil
			}
		}
	}

	return "", fmt.Errorf("must use HH:MM or HH:MM:SS format")
}

func coerceJSON(value any) (string, error) {
	switch typed := value.(type) {
	case string:
		raw := strings.TrimSpace(typed)
		if raw == "" {
			return "{}", nil
		}
		var decoded any
		if err := json.Unmarshal([]byte(raw), &decoded); err != nil {
			return "", fmt.Errorf("must be valid JSON")
		}
		return raw, nil
	default:
		bytes, err := json.Marshal(typed)
		if err != nil {
			return "", fmt.Errorf("must be valid JSON")
		}
		return string(bytes), nil
	}
}
