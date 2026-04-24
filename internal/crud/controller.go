package crud

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	legacyconfig "gogal/config"
	"gogal/controllers"
	"gogal/internal/migration"
	"gogal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetDocTypeMetaByDoctype(c *gin.Context) {
	c.Params = append(c.Params, gin.Param{Key: "name", Value: c.Param("doctype")})
	controllers.GetDocTypeMeta(c)
}

func GetResourceByIDAlias(c *gin.Context) {
	c.Params = append(c.Params, gin.Param{Key: "name", Value: c.Param("id")})
	controllers.GetResource(c)
}

func UpdateResourceByIDAlias(c *gin.Context) {
	c.Params = append(c.Params, gin.Param{Key: "name", Value: c.Param("id")})
	controllers.UpdateResource(c)
}

func DeleteResourceByIDAlias(c *gin.Context) {
	c.Params = append(c.Params, gin.Param{Key: "name", Value: c.Param("id")})
	controllers.DeleteResource(c)
}

func UpdateDocType(c *gin.Context) {
	var req models.CreateDocTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload", "details": err.Error()})
		return
	}

	name := strings.TrimSpace(c.Param("name"))
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "doctype name is required in route"})
		return
	}

	existing, err := models.LoadDocTypeByName(legacyconfig.DB, name, true)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("DocType %q not found", name)})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load doctype", "details": err.Error()})
		return
	}

	// Route param is canonical for edits.
	req.Name = existing.Name
	if strings.TrimSpace(req.Label) == "" {
		req.Label = existing.Label
	}
	if strings.TrimSpace(req.Module) == "" {
		req.Module = existing.Module
	}
	if strings.TrimSpace(req.TableName) == "" {
		req.TableName = existing.StorageTable
	}
	if strings.TrimSpace(req.TableName) != existing.StorageTable {
		c.JSON(http.StatusBadRequest, gin.H{"error": "table_name cannot be changed after creation"})
		return
	}

	normalized, err := models.NewDocTypeFromRequest(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid doctype payload", "details": err.Error()})
		return
	}

	txErr := legacyconfig.DB.Transaction(func(tx *gorm.DB) error {
		updateColumns := map[string]any{
			"label":          normalized.Label,
			"module":         normalized.Module,
			"table_name":     normalized.StorageTable,
			"description":    normalized.Description,
			"is_single":      normalized.IsSingle,
			"is_child_table": normalized.IsChildTable,
			"track_changes":  normalized.TrackChanges,
			"allow_rename":   normalized.AllowRename,
			"quick_entry":    normalized.QuickEntry,
		}
		if err := tx.Model(&models.DocType{}).Where("id = ?", existing.ID).Updates(updateColumns).Error; err != nil {
			return err
		}

		if err := tx.Unscoped().Where("doc_type_id = ?", existing.ID).Delete(&models.DocField{}).Error; err != nil {
			return err
		}

		for i := range normalized.Fields {
			f := normalized.Fields[i]
			f.DocTypeID = existing.ID
			f.SortOrder = i + 1
			if err := tx.Create(&f).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if txErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update doctype", "details": txErr.Error()})
		return
	}

	// Non-destructive schema sync: only create/add/unique.
	exec := migration.NewExecutor(legacyconfig.DB)
	if err := exec.SyncAllDocTypes(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "doctype updated but migration sync failed", "details": err.Error()})
		return
	}

	updated, err := models.LoadDocTypeByName(legacyconfig.DB, existing.Name, true)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "DocType updated", "warning": "reload failed", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "DocType updated", "data": updated})
}
