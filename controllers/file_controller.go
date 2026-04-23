package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gogal/config"
	"gogal/models"

	"github.com/gin-gonic/gin"
)

func ListFiles(c *gin.Context) {
	query := config.DB.Model(&models.File{}).Where("deleted_at IS NULL")
	if attachedToDocType := strings.TrimSpace(c.Query("attached_to_doctype")); attachedToDocType != "" {
		query = query.Where("attached_to_doc_type = ?", attachedToDocType)
	}
	if attachedToName := strings.TrimSpace(c.Query("attached_to_name")); attachedToName != "" {
		query = query.Where("attached_to_name = ?", attachedToName)
	}
	if visibility := strings.TrimSpace(c.Query("visibility")); visibility != "" {
		query = query.Where("visibility = ?", normalizeVisibility(visibility))
	}

	limit := parsePositiveInt(c.DefaultQuery("limit", "20"), 20, 100)
	files := make([]models.File, 0, limit)
	if err := query.Order("created_at DESC, id DESC").Limit(limit).Find(&files).Error; err != nil {
		log.Printf("list files failed: %v", err)
		jsonError(c, http.StatusInternalServerError, "failed_to_list_files", "Unable to list files.", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": files})
}

func UploadFile(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		jsonError(c, http.StatusBadRequest, "missing_file", "Upload requires a multipart file field named \"file\".", err.Error())
		return
	}

	visibility := normalizeVisibility(c.DefaultPostForm("visibility", "public"))
	if visibility != "public" && visibility != "private" {
		jsonError(c, http.StatusBadRequest, "invalid_visibility", "Visibility must be public or private.", nil)
		return
	}

	attributes := strings.TrimSpace(c.PostForm("attributes"))
	if attributes == "" {
		attributes = "{}"
	}
	var parsedAttributes any
	if err := json.Unmarshal([]byte(attributes), &parsedAttributes); err != nil {
		jsonError(c, http.StatusBadRequest, "invalid_attributes", "Attributes must be valid JSON.", err.Error())
		return
	}

	uploadedFile, err := fileHeader.Open()
	if err != nil {
		jsonError(c, http.StatusBadRequest, "failed_to_open_upload", "Unable to open the uploaded file.", err.Error())
		return
	}
	defer uploadedFile.Close()

	name, err := models.GenerateDocumentName("file")
	if err != nil {
		jsonError(c, http.StatusInternalServerError, "failed_to_name_file", "Unable to generate a file identifier.", err.Error())
		return
	}

	extension := strings.ToLower(filepath.Ext(fileHeader.Filename))
	storedName := fmt.Sprintf("%s%s", name, extension)
	relativeDir := filepath.ToSlash(filepath.Join(visibility, time.Now().UTC().Format("2006/01")))
	absoluteDir := filepath.Join(storageRoot(), relativeDir)
	if err := os.MkdirAll(absoluteDir, 0o755); err != nil {
		jsonError(c, http.StatusInternalServerError, "failed_to_prepare_storage", "Unable to prepare the storage directory.", err.Error())
		return
	}

	absolutePath := filepath.Join(absoluteDir, storedName)
	destination, err := os.Create(absolutePath)
	if err != nil {
		jsonError(c, http.StatusInternalServerError, "failed_to_store_file", "Unable to create the destination file.", err.Error())
		return
	}
	defer destination.Close()

	if _, err := io.Copy(destination, uploadedFile); err != nil {
		jsonError(c, http.StatusInternalServerError, "failed_to_store_file", "Unable to copy the uploaded file.", err.Error())
		return
	}

	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" && extension != "" {
		contentType = mime.TypeByExtension(extension)
	}

	relativePath := filepath.ToSlash(filepath.Join(relativeDir, storedName))
	fileRecord := &models.File{
		Name:              name,
		OriginalName:      fileHeader.Filename,
		StoredName:        storedName,
		StoragePath:       relativePath,
		FileURL:           publicFileURL(visibility, relativePath),
		Visibility:        visibility,
		ContentType:       contentType,
		Extension:         extension,
		SizeBytes:         fileHeader.Size,
		AttachedToDocType: strings.TrimSpace(c.PostForm("attached_to_doctype")),
		AttachedToName:    strings.TrimSpace(c.PostForm("attached_to_name")),
		AttachedToField:   strings.TrimSpace(c.PostForm("attached_to_field")),
		AltText:           strings.TrimSpace(c.PostForm("alt_text")),
		Attributes:        attributes,
	}

	if err := config.DB.Create(fileRecord).Error; err != nil {
		jsonError(c, http.StatusInternalServerError, "failed_to_save_file_metadata", "Unable to save file metadata.", err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": fmt.Sprintf("File %q uploaded successfully.", fileRecord.OriginalName),
		"data":    fileRecord,
	})
}

func storageRoot() string {
	if value := strings.TrimSpace(os.Getenv("GOGAL_STORAGE_ROOT")); value != "" {
		return value
	}
	return filepath.Join("storage")
}

func normalizeVisibility(value string) string {
	if strings.EqualFold(strings.TrimSpace(value), "private") {
		return "private"
	}
	return "public"
}

func publicFileURL(visibility, relativePath string) string {
	if visibility == "private" {
		return ""
	}
	trimmed := strings.TrimPrefix(relativePath, "public/")
	return "/files/" + strings.TrimPrefix(trimmed, "/")
}
