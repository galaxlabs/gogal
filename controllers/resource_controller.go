package controllers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"gogal-framework/config"
	"gogal-framework/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ListResources(c *gin.Context) {
	docType, ok := loadDocTypeForResource(c, true)
	if !ok {
		return
	}

	limit := parsePositiveInt(c.DefaultQuery("limit", "20"), 20, 100)
	offset := parsePositiveInt(c.DefaultQuery("offset", "0"), 0, 100000)

	var total int64
	if err := config.DB.Table(docType.StorageTable).Where("deleted_at IS NULL").Count(&total).Error; err != nil {
		log.Printf("count records for %s failed: %v", docType.Name, err)
		jsonError(c, http.StatusInternalServerError, "failed_to_count_records", "Unable to count records.", err.Error())
		return
	}

	records := make([]map[string]any, 0)
	if err := config.DB.Table(docType.StorageTable).
		Select(models.DocumentSelectColumns(docType)).
		Where("deleted_at IS NULL").
		Order("updated_at DESC, id DESC").
		Limit(limit).
		Offset(offset).
		Find(&records).Error; err != nil {
		log.Printf("list records for %s failed: %v", docType.Name, err)
		jsonError(c, http.StatusInternalServerError, "failed_to_list_records", "Unable to list records.", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": records,
		"meta": gin.H{
			"doctype": docType.Name,
			"limit":   limit,
			"offset":  offset,
			"total":   total,
		},
	})
}

func GetResource(c *gin.Context) {
	docType, ok := loadDocTypeForResource(c, true)
	if !ok {
		return
	}

	recordName, err := models.NormalizeDocumentName(c.Param("name"))
	if err != nil {
		jsonError(c, http.StatusBadRequest, "invalid_record_name", err.Error(), nil)
		return
	}

	record, err := fetchResourceRecord(docType, recordName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			jsonError(c, http.StatusNotFound, "record_not_found", fmt.Sprintf("Record %q was not found in %q.", recordName, docType.Name), nil)
			return
		}

		log.Printf("get record %s/%s failed: %v", docType.Name, recordName, err)
		jsonError(c, http.StatusInternalServerError, "failed_to_get_record", "Unable to load record.", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": record})
}

func CreateResource(c *gin.Context) {
	docType, ok := loadDocTypeForResource(c, true)
	if !ok {
		return
	}

	var payload map[string]any
	if err := c.ShouldBindJSON(&payload); err != nil {
		jsonError(c, http.StatusBadRequest, "invalid_payload", "Request body must be valid JSON.", err.Error())
		return
	}

	recordName, err := extractOrGenerateRecordName(docType, payload)
	if err != nil {
		jsonError(c, http.StatusBadRequest, "invalid_record_name", err.Error(), nil)
		return
	}

	delete(payload, "name")
	prepared, err := models.PrepareDocumentPayload(docType, payload, true)
	if err != nil {
		jsonError(c, http.StatusBadRequest, "invalid_record_payload", err.Error(), nil)
		return
	}

	values := make(map[string]any, len(prepared)+1)
	values["name"] = recordName
	for key, value := range prepared {
		values[key] = value
	}

	if err := config.DB.Table(docType.StorageTable).Create(values).Error; err != nil {
		log.Printf("create record in %s failed: %v", docType.Name, err)
		jsonError(c, http.StatusBadRequest, "failed_to_create_record", err.Error(), nil)
		return
	}

	created, err := fetchResourceRecord(docType, recordName)
	if err != nil {
		log.Printf("reload created record %s/%s failed: %v", docType.Name, recordName, err)
		jsonError(c, http.StatusCreated, "record_created_with_reload_warning", "Record created, but could not be reloaded cleanly.", err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": fmt.Sprintf("Record %q created successfully.", recordName),
		"data":    created,
	})
}

func UpdateResource(c *gin.Context) {
	docType, ok := loadDocTypeForResource(c, true)
	if !ok {
		return
	}

	recordName, err := models.NormalizeDocumentName(c.Param("name"))
	if err != nil {
		jsonError(c, http.StatusBadRequest, "invalid_record_name", err.Error(), nil)
		return
	}

	var payload map[string]any
	if err := c.ShouldBindJSON(&payload); err != nil {
		jsonError(c, http.StatusBadRequest, "invalid_payload", "Request body must be valid JSON.", err.Error())
		return
	}

	if _, hasName := payload["name"]; hasName {
		jsonError(c, http.StatusBadRequest, "immutable_record_name", "Field \"name\" cannot be updated once a record is created.", nil)
		return
	}

	prepared, err := models.PrepareDocumentPayload(docType, payload, false)
	if err != nil {
		jsonError(c, http.StatusBadRequest, "invalid_record_payload", err.Error(), nil)
		return
	}

	if len(prepared) == 0 {
		jsonError(c, http.StatusBadRequest, "empty_update", "Provide at least one editable field to update.", nil)
		return
	}

	prepared["updated_at"] = gorm.Expr("NOW()")
	result := config.DB.Table(docType.StorageTable).
		Where("name = ? AND deleted_at IS NULL", recordName).
		Updates(prepared)
	if result.Error != nil {
		log.Printf("update record %s/%s failed: %v", docType.Name, recordName, result.Error)
		jsonError(c, http.StatusBadRequest, "failed_to_update_record", result.Error.Error(), nil)
		return
	}

	if result.RowsAffected == 0 {
		jsonError(c, http.StatusNotFound, "record_not_found", fmt.Sprintf("Record %q was not found in %q.", recordName, docType.Name), nil)
		return
	}

	updated, err := fetchResourceRecord(docType, recordName)
	if err != nil {
		log.Printf("reload updated record %s/%s failed: %v", docType.Name, recordName, err)
		jsonError(c, http.StatusOK, "record_updated_with_reload_warning", "Record updated, but could not be reloaded cleanly.", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Record %q updated successfully.", recordName),
		"data":    updated,
	})
}

func DeleteResource(c *gin.Context) {
	docType, ok := loadDocTypeForResource(c, false)
	if !ok {
		return
	}

	recordName, err := models.NormalizeDocumentName(c.Param("name"))
	if err != nil {
		jsonError(c, http.StatusBadRequest, "invalid_record_name", err.Error(), nil)
		return
	}

	result := config.DB.Table(docType.StorageTable).
		Where("name = ? AND deleted_at IS NULL", recordName).
		Updates(map[string]any{
			"deleted_at": gorm.Expr("NOW()"),
			"updated_at": gorm.Expr("NOW()"),
		})
	if result.Error != nil {
		log.Printf("delete record %s/%s failed: %v", docType.Name, recordName, result.Error)
		jsonError(c, http.StatusInternalServerError, "failed_to_delete_record", "Unable to delete record.", result.Error.Error())
		return
	}

	if result.RowsAffected == 0 {
		jsonError(c, http.StatusNotFound, "record_not_found", fmt.Sprintf("Record %q was not found in %q.", recordName, docType.Name), nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Record %q deleted successfully.", recordName),
	})
}

func loadDocTypeForResource(c *gin.Context, withFields bool) (*models.DocType, bool) {
	docType, err := loadDocTypeMetadata(c.Param("doctype"), withFields)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			jsonError(c, http.StatusNotFound, "doctype_not_found", fmt.Sprintf("DocType %q was not found.", c.Param("doctype")), nil)
			return nil, false
		}

		log.Printf("load doctype %s failed: %v", c.Param("doctype"), err)
		jsonError(c, http.StatusInternalServerError, "failed_to_load_doctype", "Unable to load DocType metadata.", err.Error())
		return nil, false
	}

	if docType.IsSingle {
		jsonError(c, http.StatusBadRequest, "single_doctype_not_supported", fmt.Sprintf("Dynamic CRUD for single DocType %q is not implemented yet.", docType.Name), nil)
		return nil, false
	}

	return docType, true
}

func loadDocTypeMetadata(name string, withFields bool) (*models.DocType, error) {
	trimmedName := strings.TrimSpace(name)
	var docType models.DocType
	query := config.DB.Model(&models.DocType{})
	if withFields {
		query = query.Preload("Fields", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("sort_order ASC, id ASC")
		})
	}

	err := query.Where("LOWER(name) = LOWER(?)", trimmedName).First(&docType).Error
	if err != nil {
		return nil, err
	}

	return &docType, nil
}

func fetchResourceRecord(docType *models.DocType, recordName string) (map[string]any, error) {
	var record map[string]any
	err := config.DB.Table(docType.StorageTable).
		Select(models.DocumentSelectColumns(docType)).
		Where("name = ? AND deleted_at IS NULL", recordName).
		Take(&record).Error
	if err != nil {
		return nil, err
	}

	return record, nil
}

func extractOrGenerateRecordName(docType *models.DocType, payload map[string]any) (string, error) {
	rawName, hasName := payload["name"]
	if hasName {
		if rawName == nil {
			return "", fmt.Errorf("document name is required")
		}
		name, err := models.NormalizeDocumentName(fmt.Sprint(rawName))
		if err != nil {
			return "", err
		}
		return name, nil
	}

	return models.GenerateDocumentName(docType.Name)
}

func parsePositiveInt(raw string, fallback int, max int) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || parsed < 0 {
		return fallback
	}
	if parsed > max {
		return max
	}
	return parsed
}
