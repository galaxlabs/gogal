package migration

import (
	"fmt"

	"gogal/models"

	"gorm.io/gorm"
)

type PlanItemType string

const (
	PlanCreateTable PlanItemType = "create_table"
	PlanAddColumn   PlanItemType = "add_column"
	PlanAddUnique   PlanItemType = "add_unique"
)

type PlanItem struct {
	Type      PlanItemType
	DocType   string
	Table     string
	Column    string
	Statement string
}

type Plan struct {
	Items          []PlanItem
	UnsafeWarnings []string
}

func BuildPlan(db *gorm.DB) (*Plan, error) {
	var doctypes []models.DocType
	if err := db.Preload("Fields", func(tx *gorm.DB) *gorm.DB {
		return tx.Order("sort_order ASC, id ASC")
	}).Where("deleted_at IS NULL").Find(&doctypes).Error; err != nil {
		return nil, fmt.Errorf("load doctypes: %w", err)
	}

	plan := &Plan{Items: make([]PlanItem, 0), UnsafeWarnings: []string{"No drop statements are generated automatically.", "Manual review required for destructive schema changes."}}

	for i := range doctypes {
		dt := &doctypes[i]
		table := ResolveTableName(dt)
		exists, err := tableExists(db, table)
		if err != nil {
			return nil, err
		}
		if !exists {
			stmt, err := BuildCreateTableSQL(dt)
			if err != nil {
				return nil, err
			}
			plan.Items = append(plan.Items, PlanItem{Type: PlanCreateTable, DocType: dt.Name, Table: table, Statement: stmt})
		}

		for _, field := range dt.Fields {
			if !models.IsStoredInParentTable(field) {
				continue
			}
			column := models.NormalizeIdentifier(field.FieldName)
			if column == "" {
				continue
			}
			colExists, err := columnExists(db, table, column)
			if err != nil {
				return nil, err
			}
			if !colExists {
				stmt, err := BuildAddColumnSQL(table, field)
				if err != nil {
					continue
				}
				plan.Items = append(plan.Items, PlanItem{Type: PlanAddColumn, DocType: dt.Name, Table: table, Column: column, Statement: stmt})
			}
			if field.Unique {
				constraint := uniqueConstraintName(table, column)
				uniqueExists, err := constraintExists(db, table, constraint)
				if err != nil {
					return nil, err
				}
				if !uniqueExists {
					stmt := fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s UNIQUE (%s);", quoteIdent(table), quoteIdent(constraint), quoteIdent(column))
					plan.Items = append(plan.Items, PlanItem{Type: PlanAddUnique, DocType: dt.Name, Table: table, Column: column, Statement: stmt})
				}
			}
		}
	}

	return plan, nil
}

func tableExists(db *gorm.DB, table string) (bool, error) {
	var exists bool
	err := db.Raw("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema='public' AND table_name=?)", table).Scan(&exists).Error
	return exists, err
}

func columnExists(db *gorm.DB, table string, column string) (bool, error) {
	var exists bool
	err := db.Raw("SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema='public' AND table_name=? AND column_name=?)", table, column).Scan(&exists).Error
	return exists, err
}

func constraintExists(db *gorm.DB, table string, constraint string) (bool, error) {
	var exists bool
	err := db.Raw("SELECT EXISTS (SELECT 1 FROM pg_constraint c JOIN pg_class t ON c.conrelid=t.oid WHERE t.relname=? AND c.conname=?)", table, constraint).Scan(&exists).Error
	return exists, err
}

func uniqueConstraintName(table string, column string) string {
	return fmt.Sprintf("%s_%s_key", table, column)
}
