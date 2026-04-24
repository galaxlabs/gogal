package migration

import (
	"fmt"
	"strings"

	"gogal/models"
)

func quoteIdent(v string) string {
	return `"` + strings.ReplaceAll(v, `"`, `""`) + `"`
}

func SQLTypeForFieldType(fieldType string) (string, bool) {
	switch fieldType {
	case "Data", "Text", "Select", "Link", "Attach", "Attach Image", "Image", "Long Text", "Small Text":
		return "TEXT", true
	case "Int":
		return "INTEGER", true
	case "Float", "Currency", "Percent":
		return "NUMERIC(18,6)", true
	case "Date":
		return "DATE", true
	case "Datetime":
		return "TIMESTAMPTZ", true
	case "Check":
		return "BOOLEAN", true
	case "Table", "Section Break", "Column Break", "Tab Break":
		return "", false
	default:
		return "", false
	}
}

func ResolveTableName(dt *models.DocType) string {
	if strings.TrimSpace(dt.StorageTable) != "" {
		return dt.StorageTable
	}
	return models.BuildStorageTableName(dt.Name)
}

func BuildCreateTableSQL(dt *models.DocType) (string, error) {
	table := ResolveTableName(dt)
	if err := models.ValidateIdentifier(strings.TrimPrefix(table, "tab_")); err != nil {
		return "", fmt.Errorf("invalid table name: %w", err)
	}

	cols := []string{
		"id BIGSERIAL PRIMARY KEY",
		"name VARCHAR(140) NOT NULL UNIQUE",
		"created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()",
		"updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()",
		"deleted_at TIMESTAMPTZ NULL",
	}
	if dt.IsChildTable {
		cols = append(cols,
			"parent VARCHAR(140) NOT NULL",
			"parenttype VARCHAR(140) NOT NULL",
			"parentfield VARCHAR(140) NOT NULL",
			"idx INTEGER NOT NULL DEFAULT 0",
		)
	}

	for _, f := range dt.Fields {
		col := models.NormalizeIdentifier(f.FieldName)
		if col == "" || !models.IsStoredInParentTable(f) {
			continue
		}
		t, ok := SQLTypeForFieldType(f.FieldType)
		if !ok {
			continue
		}
		line := fmt.Sprintf("%s %s", quoteIdent(col), t)
		if f.Required {
			line += " NOT NULL"
		}
		if f.Unique {
			line += " UNIQUE"
		}
		cols = append(cols, line)
	}

	return fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s);", quoteIdent(table), strings.Join(cols, ", ")), nil
}

func BuildAddColumnSQL(tableName string, field models.DocField) (string, error) {
	column := models.NormalizeIdentifier(field.FieldName)
	if column == "" {
		return "", fmt.Errorf("invalid field name")
	}
	typeSQL, ok := SQLTypeForFieldType(field.FieldType)
	if !ok {
		return "", fmt.Errorf("unsupported field type %q", field.FieldType)
	}
	stmt := fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS %s %s", quoteIdent(tableName), quoteIdent(column), typeSQL)
	if field.Required {
		switch field.FieldType {
		case "Int":
			stmt += " NOT NULL DEFAULT 0"
		case "Float", "Currency", "Percent":
			stmt += " NOT NULL DEFAULT 0"
		case "Check":
			stmt += " NOT NULL DEFAULT false"
		default:
			stmt += " NOT NULL DEFAULT ''"
		}
	}
	stmt += ";"
	return stmt, nil
}
