package studio

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.Engine) {
	r.GET("/desk", DeskHome)
	r.GET("/desk/dashboard", Dashboard)
	r.GET("/desk/apps", AppsPage)
	r.GET("/desk/modules", ModulesPage)
	r.GET("/desk/doctypes", DocTypeList)
	r.GET("/desk/doctypes/new", NewDocType)
	r.GET("/desk/doctypes/:name", ViewDocType)
	r.GET("/desk/resource/:doctype", ResourceList)
	r.GET("/desk/resource/:doctype/table", ResourceTablePartial)
	r.GET("/desk/resource/:doctype/new", ResourceNew)
	r.GET("/desk/resource/:doctype/:id", ResourceEdit)
	r.GET("/desk/resource/:doctype/:id/timeline", ResourceTimelinePartial)
	r.POST("/desk/resource/:doctype/:id/comment", ResourceCommentAction)
	r.POST("/desk/resource/:doctype/:id/comment/:comment_id/edit", ResourceCommentEditPlaceholder)
	r.POST("/desk/resource/:doctype/:id/comment/:comment_id/delete", ResourceCommentDeleteAction)
	r.POST("/desk/resource/:doctype/:id/assign", ResourceAssignmentAction)
	r.POST("/desk/resource/:doctype/:id/assignment/:assignment_id/:status", ResourceAssignmentStatusAction)
	r.DELETE("/desk/resource/:doctype/:id", ResourceDeleteAction)
	r.GET("/desk/bench", BenchPage)
	r.GET("/bench", BenchManagerHome)
}
