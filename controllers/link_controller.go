package controllers

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"gogal/config"
	"gogal/models"

	"github.com/gin-gonic/gin"
)

func SearchLinkOptions(c *gin.Context) {
	docType, ok := loadDocTypeForResource(c, true)
	if !ok {
		return
	}

	limit := parsePositiveInt(c.DefaultQuery("limit", "10"), 10, 50)
	search := strings.TrimSpace(c.Query("search"))
	query := config.DB.Table(docType.StorageTable).Where("deleted_at IS NULL")
	query, _ = applyResourceSearch(query, docType, search)

	selectColumns := linkSearchSelectColumns(docType)
	records := make([]map[string]any, 0, limit)
	if err := query.Select(selectColumns).Order("updated_at DESC, id DESC").Limit(limit).Find(&records).Error; err != nil {
		log.Printf("search link options for %s failed: %v", docType.Name, err)
		jsonError(c, http.StatusInternalServerError, "failed_to_search_link_options", "Unable to load link options.", err.Error())
		return
	}

	options := make([]gin.H, 0, len(records))
	for _, record := range records {
		name := strings.TrimSpace(fmt.Sprint(record["name"]))
		if name == "" {
			continue
		}

		option := gin.H{"name": name, "label": name}
		for _, field := range docType.Fields {
			if !models.IsTextLikeFieldType(field.FieldType) || field.FieldName == "name" {
				continue
			}
			value := strings.TrimSpace(fmt.Sprint(record[field.FieldName]))
			if value == "" {
				continue
			}
			if option["label"] == name {
				option["label"] = value
			} else {
				option["description"] = value
				break
			}
		}

		options = append(options, option)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": options,
		"meta": gin.H{
			"doctype": docType.Name,
			"search":  search,
			"limit":   limit,
		},
	})
}

func linkSearchSelectColumns(docType *models.DocType) []string {
	columns := []string{"name"}
	for _, field := range docType.Fields {
		if !models.IsStoredInParentTable(field) || !models.IsTextLikeFieldType(field.FieldType) || field.FieldName == "name" {
			continue
		}
		columns = append(columns, field.FieldName)
		if len(columns) >= 4 {
			break
		}
	}
	return columns
}
