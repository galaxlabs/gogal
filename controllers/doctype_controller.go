package controllers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"gogal/config"
	"gogal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ListDocTypes(c *gin.Context) {
	var docTypes []models.DocType
	if err := config.DB.Order("name ASC").Find(&docTypes).Error; err != nil {
		log.Printf("list doctypes failed: %v", err)
		jsonError(c, http.StatusInternalServerError, "failed_to_list_doctypes", "Unable to load doctypes.", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": docTypes})
}

func GetDocTypeMeta(c *gin.Context) {
	docTypeName := strings.TrimSpace(c.Param("name"))
	docType, err := loadDocTypeMetadata(docTypeName, true)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			jsonError(c, http.StatusNotFound, "doctype_not_found", fmt.Sprintf("DocType %q was not found.", docTypeName), nil)
			return
		}

		log.Printf("load doctype %s failed: %v", docTypeName, err)
		jsonError(c, http.StatusInternalServerError, "failed_to_load_doctype", "Unable to load DocType metadata.", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": docType})
}

func CreateDocType(c *gin.Context) {
	var request models.CreateDocTypeRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		jsonError(c, http.StatusBadRequest, "invalid_payload", "Request body is invalid JSON metadata for a DocType.", err.Error())
		return
	}

	docType, err := models.NewDocTypeFromRequest(request)
	if err != nil {
		jsonError(c, http.StatusBadRequest, "invalid_doctype", err.Error(), gin.H{"supported_fieldtypes": models.SupportedFieldTypes()})
		return
	}

	err = config.DB.Transaction(func(tx *gorm.DB) error {
		var existingCount int64
		if err := tx.Model(&models.DocType{}).
			Where("name = ? OR table_name = ?", docType.Name, docType.StorageTable).
			Count(&existingCount).Error; err != nil {
			return err
		}

		if existingCount > 0 {
			return fmt.Errorf("doctype %q or table %q already exists", docType.Name, docType.StorageTable)
		}

		if err := validateDocTypeFieldReferences(tx, docType); err != nil {
			return err
		}

		if err := tx.Create(docType).Error; err != nil {
			return err
		}

		if docType.IsSingle {
			return nil
		}

		statement, err := buildCreateTableStatement(docType)
		if err != nil {
			return err
		}

		if err := tx.Exec(statement).Error; err != nil {
			return fmt.Errorf("create storage table %s: %w", docType.StorageTable, err)
		}

		return nil
	})

	if err != nil {
		log.Printf("create doctype failed: %v", err)
		jsonError(c, http.StatusBadRequest, "failed_to_create_doctype", err.Error(), nil)
		return
	}

	var created models.DocType
	if err := config.DB.Preload("Fields", func(tx *gorm.DB) *gorm.DB {
		return tx.Order("sort_order ASC, id ASC")
	}).Where("name = ?", docType.Name).First(&created).Error; err != nil {
		log.Printf("reload created doctype %s failed: %v", docType.Name, err)
		jsonError(c, http.StatusCreated, "doctype_created_with_reload_warning", "DocType created, but it could not be reloaded cleanly.", err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": fmt.Sprintf("DocType %q created successfully.", created.Name),
		"data":    created,
	})
}

func jsonError(c *gin.Context, status int, code, message string, details any) {
	response := gin.H{
		"error": gin.H{
			"code":    code,
			"message": message,
		},
	}

	if details != nil {
		response["error"].(gin.H)["details"] = details
	}

	c.JSON(status, response)
}

func buildCreateTableStatement(docType *models.DocType) (string, error) {
	if err := models.ValidateIdentifier(docType.StorageTable); err != nil {
		return "", err
	}

	columnDefinitions := []string{
		"id BIGSERIAL PRIMARY KEY",
		"name VARCHAR(140) NOT NULL UNIQUE",
		"created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()",
		"updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()",
		"deleted_at TIMESTAMPTZ NULL",
	}

	if docType.IsChildTable {
		columnDefinitions = append(columnDefinitions,
			"parent VARCHAR(140) NOT NULL",
			"parenttype VARCHAR(140) NOT NULL",
			"parentfield VARCHAR(140) NOT NULL",
			"idx INTEGER NOT NULL DEFAULT 0",
		)
	}

	for _, field := range docType.Fields {
		if !models.IsStoredInParentTable(field) {
			continue
		}

		columnType, ok := models.FieldDatabaseType(field.FieldType)
		if !ok {
			return "", fmt.Errorf("unsupported fieldtype %q for field %q", field.FieldType, field.FieldName)
		}

		column := fmt.Sprintf("%s %s", field.FieldName, columnType)
		if field.Required {
			column += " NOT NULL"
		}
		if field.Unique {
			column += " UNIQUE"
		}

		columnDefinitions = append(columnDefinitions, column)
	}

	statement := fmt.Sprintf(
		"CREATE TABLE IF NOT EXISTS %s (%s);",
		docType.StorageTable,
		strings.Join(columnDefinitions, ", "),
	)

	return statement, nil
}

func validateDocTypeFieldReferences(tx *gorm.DB, docType *models.DocType) error {
	for _, field := range docType.Fields {
		targetDocTypeName := models.ResolveTargetDocTypeName(field)
		if targetDocTypeName == "" {
			continue
		}

		switch field.FieldType {
		case "Link", "Table":
			targetDocType, err := models.LoadDocTypeByName(tx, targetDocTypeName, false)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return fmt.Errorf("field %q references DocType %q, but it does not exist yet", field.FieldName, targetDocTypeName)
				}
				return fmt.Errorf("field %q: load referenced DocType %q: %w", field.FieldName, targetDocTypeName, err)
			}

			if field.FieldType == "Table" && !targetDocType.IsChildTable {
				return fmt.Errorf("field %q references DocType %q, but that DocType is not marked as a child table", field.FieldName, targetDocType.Name)
			}
		}
	}

	return nil
}
