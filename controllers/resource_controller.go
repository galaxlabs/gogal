package controllers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"gogal/config"
	"gogal/models"

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
	filteredQuery, listMeta, err := applyResourceListOptions(config.DB.Table(docType.StorageTable).Where("deleted_at IS NULL"), docType, c.Request.URL.Query())
	if err != nil {
		jsonError(c, http.StatusBadRequest, "invalid_list_query", err.Error(), nil)
		return
	}

	var total int64
	if err := filteredQuery.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		log.Printf("count records for %s failed: %v", docType.Name, err)
		jsonError(c, http.StatusInternalServerError, "failed_to_count_records", "Unable to count records.", err.Error())
		return
	}

	records := make([]map[string]any, 0)
	if err := filteredQuery.Session(&gorm.Session{}).
		Select(models.DocumentSelectColumns(docType)).
		Order(listMeta["sort"].(string)).
		Limit(limit).
		Offset(offset).
		Find(&records).Error; err != nil {
		log.Printf("list records for %s failed: %v", docType.Name, err)
		jsonError(c, http.StatusInternalServerError, "failed_to_list_records", "Unable to list records.", err.Error())
		return
	}

	if err := hydrateChildTablesForRecords(config.DB, docType, records); err != nil {
		log.Printf("hydrate child rows for %s failed: %v", docType.Name, err)
		jsonError(c, http.StatusInternalServerError, "failed_to_list_records", "Unable to hydrate child table rows.", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": records,
		"meta": gin.H{
			"doctype": docType.Name,
			"limit":   limit,
			"offset":  offset,
			"total":   total,
			"search":  listMeta["search"],
			"sort":    listMeta["sort"],
			"filters": listMeta["filters"],
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
	prepared, err := models.PrepareDocumentMutation(config.DB, docType, payload, true)
	if err != nil {
		jsonError(c, http.StatusBadRequest, "invalid_record_payload", err.Error(), nil)
		return
	}

	values := make(map[string]any, len(prepared.Values)+1)
	values["name"] = recordName
	for key, value := range prepared.Values {
		values[key] = value
	}

	if err := config.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table(docType.StorageTable).Create(values).Error; err != nil {
			return err
		}

		if err := saveChildTables(tx, docType, recordName, prepared.ChildTables); err != nil {
			return err
		}

		return nil
	}); err != nil {
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

	prepared, err := models.PrepareDocumentMutation(config.DB, docType, payload, false)
	if err != nil {
		jsonError(c, http.StatusBadRequest, "invalid_record_payload", err.Error(), nil)
		return
	}

	if len(prepared.Values) == 0 && len(prepared.ChildTables) == 0 {
		jsonError(c, http.StatusBadRequest, "empty_update", "Provide at least one editable field to update.", nil)
		return
	}

	if err := config.DB.Transaction(func(tx *gorm.DB) error {
		rowsAffected := int64(0)
		if len(prepared.Values) > 0 {
			prepared.Values["updated_at"] = gorm.Expr("NOW()")
			result := tx.Table(docType.StorageTable).
				Where("name = ? AND deleted_at IS NULL", recordName).
				Updates(prepared.Values)
			if result.Error != nil {
				return result.Error
			}
			rowsAffected = result.RowsAffected
		}

		if len(prepared.ChildTables) > 0 {
			result := tx.Table(docType.StorageTable).
				Where("name = ? AND deleted_at IS NULL", recordName).
				Update("updated_at", gorm.Expr("NOW()"))
			if result.Error != nil {
				return result.Error
			}
			if rowsAffected == 0 {
				rowsAffected = result.RowsAffected
			}

			if err := saveChildTables(tx, docType, recordName, prepared.ChildTables); err != nil {
				return err
			}
		}

		if rowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		return nil
	}); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			jsonError(c, http.StatusNotFound, "record_not_found", fmt.Sprintf("Record %q was not found in %q.", recordName, docType.Name), nil)
			return
		}

		log.Printf("update record %s/%s failed: %v", docType.Name, recordName, err)
		jsonError(c, http.StatusBadRequest, "failed_to_update_record", err.Error(), nil)
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

	if err := config.DB.Transaction(func(tx *gorm.DB) error {
		result := tx.Table(docType.StorageTable).
			Where("name = ? AND deleted_at IS NULL", recordName).
			Updates(map[string]any{
				"deleted_at": gorm.Expr("NOW()"),
				"updated_at": gorm.Expr("NOW()"),
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		return deleteAllChildTableRows(tx, docType, recordName)
	}); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			jsonError(c, http.StatusNotFound, "record_not_found", fmt.Sprintf("Record %q was not found in %q.", recordName, docType.Name), nil)
			return
		}

		log.Printf("delete record %s/%s failed: %v", docType.Name, recordName, err)
		jsonError(c, http.StatusInternalServerError, "failed_to_delete_record", "Unable to delete record.", err.Error())
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
	return models.LoadDocTypeByName(config.DB, strings.TrimSpace(name), withFields)
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

	if err := hydrateChildTablesForRecords(config.DB, docType, []map[string]any{record}); err != nil {
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

func applyResourceListOptions(tx *gorm.DB, docType *models.DocType, queryValues url.Values) (*gorm.DB, gin.H, error) {
	filteredQuery, appliedFilters, err := applyResourceFilters(tx, docType, queryValues)
	if err != nil {
		return nil, nil, err
	}

	searchedQuery, searchTerm := applyResourceSearch(filteredQuery, docType, queryValues.Get("search"))
	sortClause, err := normalizeResourceSort(docType, queryValues.Get("sort_by"), queryValues.Get("sort_order"))
	if err != nil {
		return nil, nil, err
	}

	return searchedQuery, gin.H{
		"filters": appliedFilters,
		"search":  searchTerm,
		"sort":    sortClause,
	}, nil
}

func applyResourceFilters(tx *gorm.DB, docType *models.DocType, queryValues url.Values) (*gorm.DB, []gin.H, error) {
	appliedFilters := make([]gin.H, 0)
	filteredQuery := tx

	for key, rawValues := range queryValues {
		if !strings.HasPrefix(strings.ToLower(key), "filter_") {
			continue
		}

		fieldName, operator := parseFilterKey(key)
		field, ok := models.QueryableField(docType, fieldName)
		if !ok {
			return nil, nil, fmt.Errorf("field %q cannot be filtered", fieldName)
		}

		normalizedOperator, err := normalizeFilterOperator(operator)
		if err != nil {
			return nil, nil, fmt.Errorf("filter %q: %w", key, err)
		}

		filteredQuery, err = applyFilterCondition(filteredQuery, field, normalizedOperator, rawValues)
		if err != nil {
			return nil, nil, fmt.Errorf("filter %q: %w", key, err)
		}

		appliedFilters = append(appliedFilters, gin.H{
			"field":    field.FieldName,
			"operator": normalizedOperator,
			"value":    rawValues,
		})
	}

	return filteredQuery, appliedFilters, nil
}

func applyResourceSearch(tx *gorm.DB, docType *models.DocType, rawSearch string) (*gorm.DB, string) {
	search := strings.TrimSpace(rawSearch)
	if search == "" {
		return tx, ""
	}

	searchableColumns := models.SearchableColumns(docType)
	if len(searchableColumns) == 0 {
		return tx, search
	}

	clauses := make([]string, 0, len(searchableColumns))
	values := make([]any, 0, len(searchableColumns))
	pattern := "%" + search + "%"
	for _, column := range searchableColumns {
		clauses = append(clauses, fmt.Sprintf("CAST(%s AS TEXT) ILIKE ?", column))
		values = append(values, pattern)
	}

	return tx.Where("("+strings.Join(clauses, " OR ")+")", values...), search
}

func normalizeResourceSort(docType *models.DocType, rawSortBy, rawSortOrder string) (string, error) {
	sortBy := strings.TrimSpace(rawSortBy)
	if sortBy == "" {
		return "updated_at DESC, id DESC", nil
	}

	field, ok := models.QueryableField(docType, sortBy)
	if !ok {
		return "", fmt.Errorf("field %q cannot be used for sorting", sortBy)
	}

	if !models.IsSortableField(field) {
		return "", fmt.Errorf("field %q cannot be used for sorting", field.FieldName)
	}

	direction := strings.ToUpper(strings.TrimSpace(rawSortOrder))
	if direction == "" {
		direction = "ASC"
	}

	if direction != "ASC" && direction != "DESC" {
		return "", fmt.Errorf("sort_order must be ASC or DESC")
	}

	return fmt.Sprintf("%s %s, id DESC", field.FieldName, direction), nil
}

func parseFilterKey(rawKey string) (string, string) {
	trimmed := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(rawKey)), "filter_")
	parts := strings.SplitN(trimmed, "__", 2)
	fieldName := strings.TrimSpace(parts[0])
	operator := "eq"
	if len(parts) == 2 && strings.TrimSpace(parts[1]) != "" {
		operator = strings.TrimSpace(parts[1])
	}

	return fieldName, operator
}

func normalizeFilterOperator(operator string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(operator)) {
	case "eq", "ne", "gt", "gte", "lt", "lte", "like", "ilike", "in", "isnull":
		return strings.ToLower(strings.TrimSpace(operator)), nil
	default:
		return "", fmt.Errorf("unsupported operator %q", operator)
	}
}

func applyFilterCondition(tx *gorm.DB, field models.DocField, operator string, rawValues []string) (*gorm.DB, error) {
	if len(rawValues) == 0 {
		return nil, fmt.Errorf("requires a value")
	}

	column := field.FieldName
	combinedValue := strings.TrimSpace(rawValues[0])

	switch operator {
	case "eq":
		value, err := models.CoerceFieldValue(field, combinedValue)
		if err != nil {
			return nil, err
		}
		return tx.Where(fmt.Sprintf("%s = ?", column), value), nil
	case "ne":
		value, err := models.CoerceFieldValue(field, combinedValue)
		if err != nil {
			return nil, err
		}
		return tx.Where(fmt.Sprintf("%s <> ?", column), value), nil
	case "gt", "gte", "lt", "lte":
		value, err := models.CoerceFieldValue(field, combinedValue)
		if err != nil {
			return nil, err
		}
		comparisonOperator := map[string]string{"gt": ">", "gte": ">=", "lt": "<", "lte": "<="}[operator]
		return tx.Where(fmt.Sprintf("%s %s ?", column, comparisonOperator), value), nil
	case "like", "ilike":
		if !models.IsTextLikeFieldType(field.FieldType) && field.FieldName != "name" {
			return nil, fmt.Errorf("supports only text-like fields")
		}
		comparisonOperator := strings.ToUpper(operator)
		return tx.Where(fmt.Sprintf("CAST(%s AS TEXT) %s ?", column, comparisonOperator), "%"+combinedValue+"%"), nil
	case "in":
		items := splitAndTrimCSV(combinedValue)
		if len(items) == 0 {
			return nil, fmt.Errorf("requires a comma-separated value list")
		}

		coercedValues := make([]any, 0, len(items))
		for _, item := range items {
			value, err := models.CoerceFieldValue(field, item)
			if err != nil {
				return nil, err
			}
			coercedValues = append(coercedValues, value)
		}

		return tx.Where(fmt.Sprintf("%s IN ?", column), coercedValues), nil
	case "isnull":
		shouldBeNull, err := parseFlexibleBool(combinedValue)
		if err != nil {
			return nil, err
		}
		if shouldBeNull {
			return tx.Where(fmt.Sprintf("%s IS NULL", column)), nil
		}
		return tx.Where(fmt.Sprintf("%s IS NOT NULL", column)), nil
	default:
		return nil, fmt.Errorf("unsupported operator %q", operator)
	}
}

func splitAndTrimCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		values = append(values, trimmed)
	}
	return values
}

func parseFlexibleBool(raw string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "yes", "y", "on":
		return true, nil
	case "0", "false", "no", "n", "off":
		return false, nil
	default:
		return false, fmt.Errorf("must be a boolean")
	}
}
