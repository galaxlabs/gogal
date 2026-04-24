package studio

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	legacyconfig "gogal/config"
	"gogal/internal/benchmanager"
	"gogal/models"

	"github.com/gin-gonic/gin"
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
	docTypeName := strings.TrimSpace(c.Param("doctype"))
	dt, err := models.LoadDocTypeByName(legacyconfig.DB, docTypeName, true)
	if err != nil {
		c.HTML(http.StatusNotFound, "record_list.html", gin.H{"doctype": docTypeName, "error": err.Error()})
		return
	}
	rows := make([]map[string]any, 0)
	if legacyconfig.DB != nil {
		_ = legacyconfig.DB.Table(dt.StorageTable).Where("deleted_at IS NULL").Order("updated_at DESC").Limit(50).Find(&rows).Error
	}
	c.HTML(http.StatusOK, "record_list.html", gin.H{"doctype": dt.Name, "fields": buildDeskFields(dt.Fields), "rows": rows})
}

func ResourceNew(c *gin.Context) {
	renderResourceForm(c, "", true)
}

func ResourceEdit(c *gin.Context) {
	renderResourceForm(c, c.Param("id"), false)
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
