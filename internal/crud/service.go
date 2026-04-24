package crud

import (
	"gogal/controllers"

	"github.com/gin-gonic/gin"
)

// Reuse stable legacy controller logic while transitioning to internal package boundaries.
func RegisterAPIRoutes(rg *gin.RouterGroup) {
	rg.GET("/doctypes", controllers.ListDocTypes)
	rg.POST("/doctypes", controllers.CreateDocType)
	rg.GET("/doctypes/:name", controllers.GetDocTypeMeta)
	rg.PUT("/doctypes/:name", UpdateDocType)

	rg.GET("/meta/:doctype", GetDocTypeMetaByDoctype)

	rg.GET("/resource/:doctype", controllers.ListResources)
	rg.GET("/resource/:doctype/single", controllers.GetSingleResource)
	rg.PUT("/resource/:doctype/single", controllers.UpdateSingleResource)
	rg.DELETE("/resource/:doctype/single", controllers.DeleteSingleResource)
	rg.GET("/resource/:doctype/link-search", controllers.SearchLinkOptions)
	rg.POST("/resource/:doctype", controllers.CreateResource)
	rg.GET("/resource/:doctype/:id", GetResourceByIDAlias)
	rg.PUT("/resource/:doctype/:id", UpdateResourceByIDAlias)
	rg.DELETE("/resource/:doctype/:id", DeleteResourceByIDAlias)

	rg.GET("/files", controllers.ListFiles)
	rg.POST("/files/upload", controllers.UploadFile)
}
