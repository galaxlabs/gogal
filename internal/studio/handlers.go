package studio

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	legacyconfig "gogal/config"
	"gogal/internal/activity"
	"gogal/internal/audit"
	"gogal/internal/benchmanager"
	"gogal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type deskField struct {
	FieldName    string
	Label        string
	FieldType    string
	Required     bool
	ReadOnly     bool
	Options      []string
	RawOptions   string
	DefaultValue string
}

type recordListData struct {
	DocType   string
	Fields    []deskField
	Rows      []map[string]any
	Search    string
	SortBy    string
	SortOrder string
	Limit     int
	Error     string
}

type timelineEvent struct {
	ID    uint
	Title string
	Time  time.Time
	Note  string
	Actor string
	Kind  string
	Badge string
	Muted string
	Diffs []activity.FieldDiff
}

func DeskHome(c *gin.Context) {
	c.HTML(http.StatusOK, "layout.html", gin.H{"title": "Gogal Desk", "current_path": "/desk", "start_path": "/desk/dashboard"})
}

func Dashboard(c *gin.Context) {
	c.HTML(http.StatusOK, "dashboard.html", gin.H{"title": "Gogal Desk Dashboard"})
}

func AppsPage(c *gin.Context) {
	apps, _ := listApps()
	c.HTML(http.StatusOK, "apps.html", gin.H{"title": "Gogal Apps", "apps": apps})
}

func ModulesPage(c *gin.Context) {
	modules, _ := listModules()
	c.HTML(http.StatusOK, "modules.html", gin.H{"title": "Gogal Modules", "modules": modules})
}

func DocTypeList(c *gin.Context) {
	var doctypes []models.DocType
	if legacyconfig.DB != nil {
		_ = legacyconfig.DB.Order("name ASC").Find(&doctypes).Error
	}
	c.HTML(http.StatusOK, "doctype_list.html", gin.H{"title": "DocTypes", "doctypes": doctypes})
}

func NewDocType(c *gin.Context) {
	c.HTML(http.StatusOK, "doctype_builder.html", gin.H{"title": "New DocType", "doctype_json": "{}"})
}

func ViewDocType(c *gin.Context) {
	dtName := strings.TrimSpace(c.Param("name"))
	dt, err := models.LoadDocTypeByName(legacyconfig.DB, dtName, true)
	if err != nil {
		c.HTML(http.StatusNotFound, "doctype_builder.html", gin.H{"title": "DocType Not Found", "error": err.Error(), "doctype_json": "{}"})
		return
	}
	c.HTML(http.StatusOK, "doctype_builder.html", gin.H{"title": "Edit DocType", "doctype": dt, "doctype_json": mustJSON(dt)})
}

func ResourceList(c *gin.Context) {
	data := loadRecordListData(c)
	if data.Error != "" {
		c.HTML(http.StatusNotFound, "record_list.html", gin.H{"doctype": data.DocType, "error": data.Error})
		return
	}
	c.HTML(http.StatusOK, "record_list.html", gin.H{
		"doctype":    data.DocType,
		"fields":     data.Fields,
		"rows":       data.Rows,
		"search":     data.Search,
		"sort_by":    data.SortBy,
		"sort_order": data.SortOrder,
		"limit":      data.Limit,
	})
}

func ResourceTablePartial(c *gin.Context) {
	data := loadRecordListData(c)
	status := http.StatusOK
	if data.Error != "" {
		status = http.StatusNotFound
	}
	c.HTML(status, "record_list_table.html", gin.H{
		"doctype":    data.DocType,
		"fields":     data.Fields,
		"rows":       data.Rows,
		"search":     data.Search,
		"sort_by":    data.SortBy,
		"sort_order": data.SortOrder,
		"limit":      data.Limit,
		"error":      data.Error,
	})
}

func loadRecordListData(c *gin.Context) recordListData {
	docTypeName := strings.TrimSpace(c.Param("doctype"))
	dt, err := models.LoadDocTypeByName(legacyconfig.DB, docTypeName, true)
	if err != nil {
		return recordListData{DocType: docTypeName, Error: err.Error()}
	}

	fields := buildDeskFields(dt.Fields)
	rows := make([]map[string]any, 0)
	search := strings.TrimSpace(c.Query("search"))
	sortBy := models.NormalizeIdentifier(c.DefaultQuery("sort_by", "updated_at"))
	sortOrder := strings.ToUpper(strings.TrimSpace(c.DefaultQuery("sort_order", "DESC")))
	if sortOrder != "ASC" {
		sortOrder = "DESC"
	}
	limit := parsePositiveInt(c.DefaultQuery("limit", "50"), 50, 200)

	if legacyconfig.DB != nil {
		query := legacyconfig.DB.Table(dt.StorageTable).Where("deleted_at IS NULL")
		if search != "" {
			query = applyDeskSearch(query, dt, search)
		}
		if !isDeskSortableField(dt, sortBy) {
			sortBy = "updated_at"
		}
		_ = query.Order(fmt.Sprintf("%s %s", quoteDeskIdent(sortBy), sortOrder)).Limit(limit).Find(&rows).Error
	}

	return recordListData{
		DocType:   dt.Name,
		Fields:    fields,
		Rows:      rows,
		Search:    search,
		SortBy:    sortBy,
		SortOrder: sortOrder,
		Limit:     limit,
	}
}

func ResourceNew(c *gin.Context) {
	renderResourceForm(c, "", true)
}

func ResourceEdit(c *gin.Context) {
	renderResourceForm(c, c.Param("id"), false)
}

func ResourceTimelinePartial(c *gin.Context) {
	events := []timelineEvent{}
	recordID := strings.TrimSpace(c.Param("id"))
	docTypeName := strings.TrimSpace(c.Param("doctype"))

	if recordID == "" || recordID == "new" {
		c.HTML(http.StatusOK, "timeline.html", gin.H{"events": events, "doctype": docTypeName, "record_id": recordID})
		return
	}

	if items, err := activity.ListTimeline(legacyconfig.DB, docTypeName, recordID, 30); err == nil {
		for _, row := range items {
			events = append(events, timelineEvent{
				ID:    row.ID,
				Title: row.Title,
				Time:  row.Time,
				Note:  row.Note,
				Actor: row.Actor,
				Kind:  row.Kind,
				Badge: row.Badge,
				Muted: row.Muted,
				Diffs: row.Diffs,
			})
		}
	}

	if len(events) == 0 && legacyconfig.DB != nil {
		if dt, err := models.LoadDocTypeByName(legacyconfig.DB, docTypeName, false); err == nil {
			if dt.TrackChanges {
				record := map[string]any{}
				_ = legacyconfig.DB.Table(dt.StorageTable).
					Select("name", "created_at", "updated_at").
					Where("name = ?", recordID).
					Take(&record).Error
				if len(record) > 0 {
					events = append(events, timelineEvent{Title: "Created", Time: coerceTimelineTime(record["created_at"]), Note: "Imported from record timestamp.", Actor: "system", Kind: activity.KindAudit, Badge: "Change"})
				}
			}
		}
	}

	c.HTML(http.StatusOK, "timeline.html", gin.H{"events": events, "doctype": docTypeName, "record_id": recordID})
}

func ResourceCommentAction(c *gin.Context) {
	docTypeName := strings.TrimSpace(c.Param("doctype"))
	recordID := strings.TrimSpace(c.Param("id"))
	if !ensureTimelineTarget(c, docTypeName, recordID) {
		return
	}
	if err := activity.AddComment(legacyconfig.DB, docTypeName, recordID, deskActor(c), c.PostForm("comment")); err != nil {
		c.HTML(http.StatusBadRequest, "partials/alert.html", gin.H{"type": "error", "message": err.Error()})
		return
	}
	ResourceTimelinePartial(c)
}

func ResourceCommentDeleteAction(c *gin.Context) {
	docTypeName := strings.TrimSpace(c.Param("doctype"))
	recordID := strings.TrimSpace(c.Param("id"))
	if !ensureTimelineTarget(c, docTypeName, recordID) {
		return
	}
	commentID, err := parseUintParam(c.Param("comment_id"))
	if err != nil {
		c.HTML(http.StatusBadRequest, "partials/alert.html", gin.H{"type": "error", "message": err.Error()})
		return
	}
	if err := activity.DeleteComment(legacyconfig.DB, docTypeName, recordID, commentID); err != nil {
		c.HTML(http.StatusBadRequest, "partials/alert.html", gin.H{"type": "error", "message": err.Error()})
		return
	}
	ResourceTimelinePartial(c)
}

func ResourceCommentEditPlaceholder(c *gin.Context) {
	c.HTML(http.StatusOK, "partials/alert.html", gin.H{"type": "info", "message": "Inline comment editing placeholder is wired. Full edit modal comes after core permissions."})
}

func ResourceAssignmentAction(c *gin.Context) {
	docTypeName := strings.TrimSpace(c.Param("doctype"))
	recordID := strings.TrimSpace(c.Param("id"))
	if !ensureTimelineTarget(c, docTypeName, recordID) {
		return
	}
	if err := activity.AddAssignment(legacyconfig.DB, docTypeName, recordID, c.PostForm("assigned_to"), deskActor(c), c.PostForm("note")); err != nil {
		c.HTML(http.StatusBadRequest, "partials/alert.html", gin.H{"type": "error", "message": err.Error()})
		return
	}
	ResourceTimelinePartial(c)
}

func ResourceAssignmentStatusAction(c *gin.Context) {
	docTypeName := strings.TrimSpace(c.Param("doctype"))
	recordID := strings.TrimSpace(c.Param("id"))
	if !ensureTimelineTarget(c, docTypeName, recordID) {
		return
	}
	assignmentID, err := parseUintParam(c.Param("assignment_id"))
	if err != nil {
		c.HTML(http.StatusBadRequest, "partials/alert.html", gin.H{"type": "error", "message": err.Error()})
		return
	}
	if err := activity.SetAssignmentStatus(legacyconfig.DB, docTypeName, recordID, assignmentID, c.Param("status")); err != nil {
		c.HTML(http.StatusBadRequest, "partials/alert.html", gin.H{"type": "error", "message": err.Error()})
		return
	}
	ResourceTimelinePartial(c)
}

func ResourceDeleteAction(c *gin.Context) {
	docTypeName := strings.TrimSpace(c.Param("doctype"))
	recordID := strings.TrimSpace(c.Param("id"))
	dt, err := models.LoadDocTypeByName(legacyconfig.DB, docTypeName, false)
	if err != nil {
		c.HTML(http.StatusNotFound, "partials/alert.html", gin.H{"type": "error", "message": err.Error()})
		return
	}
	if recordID == "" {
		c.HTML(http.StatusBadRequest, "partials/alert.html", gin.H{"type": "error", "message": "record id is required"})
		return
	}
	err = legacyconfig.DB.Table(dt.StorageTable).
		Where("name = ? AND deleted_at IS NULL", recordID).
		Updates(map[string]any{"deleted_at": time.Now().UTC(), "updated_at": time.Now().UTC()}).Error
	if err != nil {
		c.HTML(http.StatusInternalServerError, "partials/alert.html", gin.H{"type": "error", "message": err.Error()})
		return
	}
	if dt.TrackChanges {
		_ = audit.Record(legacyconfig.DB, audit.EventInput{
			DocType: dt.Name,
			DocName: recordID,
			Action:  audit.ActionDelete,
			Actor:   deskActor(c),
			Summary: fmt.Sprintf("Deleted %s %s from Desk", dt.Name, recordID),
		})
	}
	c.HTML(http.StatusOK, "partials/alert.html", gin.H{"type": "success", "message": "Record deleted."})
}

func ensureTimelineTarget(c *gin.Context, docTypeName string, recordID string) bool {
	if recordID == "" || recordID == "new" {
		c.HTML(http.StatusBadRequest, "partials/alert.html", gin.H{"type": "error", "message": "Save the record before adding activity."})
		return false
	}
	dt, err := models.LoadDocTypeByName(legacyconfig.DB, docTypeName, false)
	if err != nil {
		c.HTML(http.StatusNotFound, "partials/alert.html", gin.H{"type": "error", "message": err.Error()})
		return false
	}
	var count int64
	if err := legacyconfig.DB.Table(dt.StorageTable).Where("name = ? AND deleted_at IS NULL", recordID).Count(&count).Error; err != nil || count == 0 {
		c.HTML(http.StatusNotFound, "partials/alert.html", gin.H{"type": "error", "message": "Record not found."})
		return false
	}
	return true
}

func deskActor(c *gin.Context) string {
	for _, key := range []string{"X-Gogal-User", "X-User", "X-Actor"} {
		if value := strings.TrimSpace(c.GetHeader(key)); value != "" {
			return value
		}
	}
	return "system"
}

func coerceTimelineTime(value any) time.Time {
	switch t := value.(type) {
	case time.Time:
		return t
	case string:
		if parsed, err := time.Parse(time.RFC3339, t); err == nil {
			return parsed
		}
	}
	return time.Now().UTC()
}

func parseUintParam(raw string) (uint, error) {
	parsed, err := strconv.ParseUint(strings.TrimSpace(raw), 10, 64)
	if err != nil || parsed == 0 {
		return 0, fmt.Errorf("valid id is required")
	}
	return uint(parsed), nil
}

func renderResourceForm(c *gin.Context, recordID string, isNew bool) {
	docTypeName := strings.TrimSpace(c.Param("doctype"))
	dt, err := models.LoadDocTypeByName(legacyconfig.DB, docTypeName, true)
	if err != nil {
		c.HTML(http.StatusNotFound, "record_form.html", gin.H{"doctype": docTypeName, "error": err.Error()})
		return
	}
	record := map[string]any{}
	if !isNew && recordID != "" {
		_ = legacyconfig.DB.Table(dt.StorageTable).Where("name = ? AND deleted_at IS NULL", recordID).Take(&record).Error
	}

	c.HTML(http.StatusOK, "record_form.html", gin.H{
		"doctype":   dt.Name,
		"fields":    buildDeskFields(dt.Fields),
		"record":    record,
		"record_id": recordID,
		"is_new":    isNew,
	})
}

func BenchPage(c *gin.Context) {
	status := benchmanager.LoadStatus(".")
	c.HTML(http.StatusOK, "bench_dashboard.html", gin.H{"title": "Gogal Bench Manager", "status": status})
}

func BenchManagerHome(c *gin.Context) {
	c.HTML(http.StatusOK, "layout.html", gin.H{"title": "Gogal Bench Manager", "current_path": "/bench", "start_path": "/desk/bench"})
}

func buildDeskFields(fields []models.DocField) []deskField {
	result := make([]deskField, 0, len(fields))
	sort.SliceStable(fields, func(i, j int) bool { return fields[i].SortOrder < fields[j].SortOrder })
	for _, f := range fields {
		if !models.IsStoredInParentTable(f) {
			continue
		}
		result = append(result, deskField{
			FieldName:    f.FieldName,
			Label:        f.Label,
			FieldType:    f.FieldType,
			Required:     f.Required,
			ReadOnly:     f.ReadOnly,
			Options:      parseOptions(f.Options),
			RawOptions:   f.Options,
			DefaultValue: f.DefaultValue,
		})
	}
	return result
}

func parseOptions(raw string) []string {
	parts := strings.Split(raw, "\n")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		v := strings.TrimSpace(p)
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}

func titleAction(action string) string {
	action = strings.TrimSpace(action)
	if action == "" {
		return "Event"
	}
	return strings.ToUpper(action[:1]) + action[1:]
}

func applyDeskSearch(query *gorm.DB, dt *models.DocType, search string) *gorm.DB {
	pattern := "%" + strings.ToLower(search) + "%"
	clauses := []string{"LOWER(name) LIKE ?"}
	args := []any{pattern}
	for _, field := range dt.Fields {
		if !models.IsStoredInParentTable(field) || !models.IsTextLikeFieldType(field.FieldType) {
			continue
		}
		col := models.NormalizeIdentifier(field.FieldName)
		if col == "" {
			continue
		}
		clauses = append(clauses, fmt.Sprintf("LOWER(%s) LIKE ?", quoteDeskIdent(col)))
		args = append(args, pattern)
	}
	return query.Where(strings.Join(clauses, " OR "), args...)
}

func isDeskSortableField(dt *models.DocType, fieldName string) bool {
	switch fieldName {
	case "name", "created_at", "updated_at":
		return true
	}
	for _, field := range dt.Fields {
		if field.FieldName == fieldName && models.IsStoredInParentTable(field) && models.IsSortableField(field) {
			return true
		}
	}
	return false
}

func quoteDeskIdent(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}

func parsePositiveInt(raw string, fallback int, max int) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || parsed <= 0 {
		return fallback
	}
	if parsed > max {
		return max
	}
	return parsed
}

func listApps() ([]string, error) {
	entries, err := os.ReadDir("apps")
	if err != nil {
		return nil, err
	}
	apps := make([]string, 0)
	for _, e := range entries {
		if e.IsDir() {
			apps = append(apps, e.Name())
		}
	}
	sort.Strings(apps)
	return apps, nil
}

func listModules() ([]string, error) {
	apps, err := listApps()
	if err != nil {
		return nil, err
	}
	seen := map[string]struct{}{}
	for _, app := range apps {
		modDir := filepath.Join("apps", app, "modules")
		entries, readErr := os.ReadDir(modDir)
		if readErr != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				seen[e.Name()] = struct{}{}
			}
		}
	}
	modules := make([]string, 0, len(seen))
	for m := range seen {
		modules = append(modules, m)
	}
	sort.Strings(modules)
	return modules, nil
}

func mustJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}
