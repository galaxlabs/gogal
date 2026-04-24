package activity

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"gogal/internal/audit"
	"gogal/models"

	"gorm.io/gorm"
)

const (
	KindAudit      = "audit"
	KindComment    = "comment"
	KindAssignment = "assignment"
)

type TimelineItem struct {
	ID         uint
	Kind       string
	Title      string
	Time       time.Time
	Actor      string
	Note       string
	Badge      string
	Muted      string
	Diffs      []FieldDiff
	Comment    *models.Comment
	Assignment *models.Assignment
}

type FieldDiff struct {
	Field  string `json:"field"`
	Before string `json:"before"`
	After  string `json:"after"`
}

func AddComment(db *gorm.DB, docType string, docName string, author string, body string) error {
	docType = strings.TrimSpace(docType)
	docName = strings.TrimSpace(docName)
	author = fallbackActor(author)
	body = strings.TrimSpace(body)
	if docType == "" || docName == "" {
		return fmt.Errorf("comment requires doctype and document name")
	}
	if body == "" {
		return fmt.Errorf("comment cannot be empty")
	}
	return db.Create(&models.Comment{DocType: docType, DocName: docName, Author: author, Body: body}).Error
}

func DeleteComment(db *gorm.DB, docType string, docName string, commentID uint) error {
	if db == nil {
		return nil
	}
	if commentID == 0 {
		return fmt.Errorf("comment id is required")
	}
	result := db.Where("id = ? AND doc_type = ? AND doc_name = ?", commentID, strings.TrimSpace(docType), strings.TrimSpace(docName)).
		Delete(&models.Comment{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("comment not found")
	}
	return nil
}

func AddAssignment(db *gorm.DB, docType string, docName string, assignedTo string, assignedBy string, note string) error {
	docType = strings.TrimSpace(docType)
	docName = strings.TrimSpace(docName)
	assignedTo = strings.TrimSpace(assignedTo)
	assignedBy = fallbackActor(assignedBy)
	if docType == "" || docName == "" {
		return fmt.Errorf("assignment requires doctype and document name")
	}
	if assignedTo == "" {
		return fmt.Errorf("assigned user is required")
	}
	return db.Create(&models.Assignment{
		DocType:    docType,
		DocName:    docName,
		AssignedTo: assignedTo,
		AssignedBy: assignedBy,
		Status:     "open",
		Note:       strings.TrimSpace(note),
	}).Error
}

func SetAssignmentStatus(db *gorm.DB, docType string, docName string, assignmentID uint, status string) error {
	if db == nil {
		return nil
	}
	status = strings.TrimSpace(status)
	if status != "open" && status != "closed" {
		return fmt.Errorf("assignment status must be open or closed")
	}
	result := db.Model(&models.Assignment{}).
		Where("id = ? AND doc_type = ? AND doc_name = ?", assignmentID, strings.TrimSpace(docType), strings.TrimSpace(docName)).
		Updates(map[string]any{"status": status})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("assignment not found")
	}
	return nil
}

func ListTimeline(db *gorm.DB, docType string, docName string, limit int) ([]TimelineItem, error) {
	if db == nil {
		return []TimelineItem{}, nil
	}
	if limit <= 0 || limit > 100 {
		limit = 30
	}

	items := make([]TimelineItem, 0, limit)
	logs, err := audit.ListForDocument(db, docType, docName, limit)
	if err != nil {
		return nil, err
	}
	for _, row := range logs {
		diffs := auditDiffs(row.Metadata)
		items = append(items, TimelineItem{
			ID:    row.ID,
			Kind:  KindAudit,
			Title: titleAction(row.Action),
			Time:  row.CreatedAt,
			Actor: row.Actor,
			Note:  row.Summary,
			Badge: "Change",
			Diffs: diffs,
		})
	}

	var comments []models.Comment
	if err := db.Where("doc_type = ? AND doc_name = ? AND deleted_at IS NULL", strings.TrimSpace(docType), strings.TrimSpace(docName)).
		Order("created_at DESC, id DESC").
		Limit(limit).
		Find(&comments).Error; err != nil {
		return nil, err
	}
	for _, row := range comments {
		comment := row
		items = append(items, TimelineItem{
			ID:      row.ID,
			Kind:    KindComment,
			Title:   "Comment",
			Time:    row.CreatedAt,
			Actor:   row.Author,
			Note:    row.Body,
			Badge:   "Comment",
			Comment: &comment,
		})
	}

	var assignments []models.Assignment
	if err := db.Where("doc_type = ? AND doc_name = ? AND deleted_at IS NULL", strings.TrimSpace(docType), strings.TrimSpace(docName)).
		Order("created_at DESC, id DESC").
		Limit(limit).
		Find(&assignments).Error; err != nil {
		return nil, err
	}
	for _, row := range assignments {
		assignment := row
		note := fmt.Sprintf("Assigned to %s", row.AssignedTo)
		if row.Note != "" {
			note += ": " + row.Note
		}
		items = append(items, TimelineItem{
			ID:         row.ID,
			Kind:       KindAssignment,
			Title:      "Assignment",
			Time:       row.CreatedAt,
			Actor:      row.AssignedBy,
			Note:       note,
			Badge:      titleAction(row.Status),
			Muted:      row.Status,
			Assignment: &assignment,
		})
	}

	sort.SliceStable(items, func(i, j int) bool {
		return items[i].Time.After(items[j].Time)
	})
	if len(items) > limit {
		items = items[:limit]
	}
	return items, nil
}

func auditDiffs(raw string) []FieldDiff {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "{}" {
		return nil
	}
	var meta struct {
		Diffs []FieldDiff `json:"diffs"`
	}
	if err := json.Unmarshal([]byte(raw), &meta); err != nil {
		return nil
	}
	return meta.Diffs
}

func fallbackActor(actor string) string {
	actor = strings.TrimSpace(actor)
	if actor == "" {
		return "system"
	}
	return actor
}

func titleAction(action string) string {
	action = strings.TrimSpace(action)
	if action == "" {
		return "Event"
	}
	return strings.ToUpper(action[:1]) + action[1:]
}
