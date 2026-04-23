package controllers

import (
	"fmt"

	"gogal/models"

	"gorm.io/gorm"
)

func hydrateChildTablesForRecords(db *gorm.DB, docType *models.DocType, records []map[string]any) error {
	if len(records) == 0 {
		return nil
	}

	for _, record := range records {
		parentName, _ := record["name"].(string)
		if parentName == "" {
			continue
		}

		for _, field := range docType.Fields {
			if field.FieldType != "Table" {
				continue
			}

			rows, err := fetchChildRowsForField(db, docType, field, parentName)
			if err != nil {
				return fmt.Errorf("field %q: %w", field.FieldName, err)
			}
			record[field.FieldName] = rows
		}
	}

	return nil
}

func fetchChildRowsForField(db *gorm.DB, parentDocType *models.DocType, field models.DocField, parentName string) ([]map[string]any, error) {
	childDocType, err := models.LoadDocTypeByName(db, models.ResolveTargetDocTypeName(field), true)
	if err != nil {
		return nil, err
	}

	rows := make([]map[string]any, 0)
	if err := db.Table(childDocType.StorageTable).
		Select(models.DocumentSelectColumns(childDocType)).
		Where("parent = ? AND parenttype = ? AND parentfield = ? AND deleted_at IS NULL", parentName, parentDocType.Name, field.FieldName).
		Order("idx ASC, id ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	return rows, nil
}

func saveChildTables(tx *gorm.DB, parentDocType *models.DocType, parentName string, childTables []models.PreparedChildTable) error {
	for _, childTable := range childTables {
		if err := deleteChildTableRows(tx, childTable.ChildDocType, parentDocType.Name, parentName, childTable.Field.FieldName); err != nil {
			return err
		}

		for index, row := range childTable.Rows {
			values := make(map[string]any, len(row)+4)
			for key, value := range row {
				values[key] = value
			}
			values["parent"] = parentName
			values["parenttype"] = parentDocType.Name
			values["parentfield"] = childTable.Field.FieldName
			values["idx"] = index + 1

			if err := tx.Table(childTable.ChildDocType.StorageTable).Create(values).Error; err != nil {
				return fmt.Errorf("insert child row %d into %s: %w", index+1, childTable.ChildDocType.Name, err)
			}
		}
	}

	return nil
}

func deleteAllChildTableRows(tx *gorm.DB, parentDocType *models.DocType, parentName string) error {
	for _, field := range docTypeTableFields(parentDocType) {
		childDocType, err := models.LoadDocTypeByName(tx, models.ResolveTargetDocTypeName(field), false)
		if err != nil {
			return err
		}

		if err := deleteChildTableRows(tx, childDocType, parentDocType.Name, parentName, field.FieldName); err != nil {
			return err
		}
	}

	return nil
}

func deleteChildTableRows(tx *gorm.DB, childDocType *models.DocType, parentType, parentName, parentField string) error {
	statement := fmt.Sprintf("DELETE FROM %s WHERE parent = ? AND parenttype = ? AND parentfield = ?", childDocType.StorageTable)
	if err := tx.Exec(statement, parentName, parentType, parentField).Error; err != nil {
		return fmt.Errorf("delete child rows from %s: %w", childDocType.StorageTable, err)
	}

	return nil
}

func docTypeTableFields(docType *models.DocType) []models.DocField {
	fields := make([]models.DocField, 0)
	for _, field := range docType.Fields {
		if field.FieldType == "Table" {
			fields = append(fields, field)
		}
	}
	return fields
}
