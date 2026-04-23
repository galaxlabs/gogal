package main

import (
	"log"
	"net/http"

	"gogal-framework/config"
	"gogal-framework/controllers"

	"github.com/gin-gonic/gin"
)

func main() {
	if err := config.ConnectDB(); err != nil {
		log.Fatalf("failed to initialize application: %v", err)
	}

	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Welcome to Gogal Framework API",
			"status":  "active",
		})
	})

	api := r.Group("/api")
	{
		api.GET("/doctypes", controllers.ListDocTypes)
		api.POST("/doctypes", controllers.CreateDocType)
		api.GET("/doctypes/:name/meta", controllers.GetDocTypeMeta)
		api.GET("/resource-meta/:name", controllers.GetDocTypeMeta)
		api.GET("/resource/:doctype", controllers.ListResources)
		api.POST("/resource/:doctype", controllers.CreateResource)
		api.GET("/resource/:doctype/:name", controllers.GetResource)
		api.PUT("/resource/:doctype/:name", controllers.UpdateResource)
		api.DELETE("/resource/:doctype/:name", controllers.DeleteResource)
	}

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
