package audit

import (
	"encoding/json"
	"fmt"
	"strings"

	"gogal/models"

	"gorm.io/gorm"
)

const (
	ActionCreate = "create"
	ActionUpdate = "update"
	ActionDelete = "delete"
	ActionSave   = "save"
)

type EventInput struct {
	DocType  string
	DocName  string
	Action   string
	Actor    string
	Summary  string
	Metadata map[string]any
}

func Record(db *gorm.DB, input EventInput) error {
	if db == nil {
		return nil
	}

	docType := strings.TrimSpace(input.DocType)
	docName := strings.TrimSpace(input.DocName)
	action := strings.TrimSpace(input.Action)
	if docType == "" || docName == "" || action == "" {
		return fmt.Errorf("audit event requires doctype, docname, and action")
	}

	actor := strings.TrimSpace(input.Actor)
	if actor == "" {
		actor = "system"
	}

	metadata := "{}"
	if len(input.Metadata) > 0 {
		bytes, err := json.Marshal(input.Metadata)
		if err != nil {
			return fmt.Errorf("marshal audit metadata: %w", err)
		}
		metadata = string(bytes)
	}

	event := models.AuditLog{
		DocType:  docType,
		DocName:  docName,
		Action:   action,
		Actor:    actor,
		Summary:  strings.TrimSpace(input.Summary),
		Metadata: metadata,
	}
	return db.Create(&event).Error
}

func ListForDocument(db *gorm.DB, docType string, docName string, limit int) ([]models.AuditLog, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	var rows []models.AuditLog
	if db == nil {
		return rows, nil
	}
	err := db.Where("doc_type = ? AND doc_name = ?", strings.TrimSpace(docType), strings.TrimSpace(docName)).
		Order("created_at DESC, id DESC").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}
