package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"gogal/config"
	"gogal/controllers"

	"github.com/gin-gonic/gin"
)

func main() {
	if err := config.ConnectDB(); err != nil {
		log.Fatalf("failed to initialize application: %v", err)
	}

	r := gin.Default()
	r.LoadHTMLGlob("views/*.html")
	r.Static("/files", "./storage/public")

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Welcome to Gogal API",
			"status":  "active",
		})
	})

	r.GET("/desk", func(c *gin.Context) {
		c.HTML(http.StatusOK, "layout.html", gin.H{
			"title":        "Gogal Studio",
			"current_path": "/desk",
		})
	})

	r.GET("/desk/dashboard", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`
		<section class="space-y-6">
		  <div class="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm">
		    <div class="flex flex-wrap items-center justify-between gap-3">
		      <div>
		        <p class="text-xs font-semibold uppercase tracking-[0.3em] text-sky-600">Desk</p>
		        <h2 class="mt-2 text-2xl font-semibold text-slate-900">Welcome to Gogal Studio</h2>
		        <p class="mt-2 max-w-2xl text-sm leading-6 text-slate-600">Server-rendered desk shell powered by Go templates, HTMX, and vanilla JavaScript. Use the sidebar to open the DocType builder without reloading the page.</p>
		      </div>
		      <button
		        class="inline-flex items-center rounded-2xl bg-slate-900 px-4 py-2.5 text-sm font-semibold text-white shadow-sm transition hover:bg-slate-700"
		        hx-get="/desk/doctype-builder"
		        hx-target="#app-content"
		        hx-swap="innerHTML"
		      >
		        Open DocType Builder
		      </button>
		    </div>
		  </div>

		  <div class="grid gap-4 md:grid-cols-3">
		    <div class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
		      <p class="text-xs font-semibold uppercase tracking-[0.24em] text-slate-400">Rendering</p>
		      <p class="mt-3 text-lg font-semibold text-slate-900">Go Templates</p>
		      <p class="mt-2 text-sm text-slate-600">The desk shell and builder load as HTML snippets from Gin routes.</p>
		    </div>
		    <div class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
		      <p class="text-xs font-semibold uppercase tracking-[0.24em] text-slate-400">Interactivity</p>
		      <p class="mt-3 text-lg font-semibold text-slate-900">HTMX</p>
		      <p class="mt-2 text-sm text-slate-600">Sidebar navigation and form actions swap only the content area.</p>
		    </div>
		    <div class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
		      <p class="text-xs font-semibold uppercase tracking-[0.24em] text-slate-400">Behavior</p>
		      <p class="mt-3 text-lg font-semibold text-slate-900">Vanilla JS</p>
		      <p class="mt-2 text-sm text-slate-600">Small scripts manage field selection, ordering, and preview payload generation.</p>
		    </div>
		  </div>
		</section>`))
	})

	r.GET("/desk/doctype-builder", func(c *gin.Context) {
		c.HTML(http.StatusOK, "doctype_builder.html", gin.H{
			"title": "DocType Builder",
		})
	})

	r.POST("/desk/doctype-builder", func(c *gin.Context) {
		doctypeName := strings.TrimSpace(c.PostForm("doctype"))
		label := strings.TrimSpace(c.PostForm("label"))
		moduleName := strings.TrimSpace(c.PostForm("module"))
		fieldsJSON := strings.TrimSpace(c.PostForm("fields_json"))

		if doctypeName == "" {
			c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(`<div class="rounded-2xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700">DocType name is required before saving.</div>`))
			return
		}

		if label == "" {
			label = doctypeName
		}

		if moduleName == "" {
			moduleName = "Core"
		}

		message := fmt.Sprintf("Draft DocType <strong>%s</strong> for module <strong>%s</strong> captured successfully. HTMX can now swap this response with a saved-state preview or success alert. Current field payload length: %d bytes.", label, moduleName, len(fieldsJSON))
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(fmt.Sprintf(`<div class="rounded-2xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700">%s</div>`, message)))
	})

	api := r.Group("/api")
	{
		api.GET("/doctypes", controllers.ListDocTypes)
		api.POST("/doctypes", controllers.CreateDocType)
		api.GET("/doctypes/:name/meta", controllers.GetDocTypeMeta)
		api.GET("/resource-meta/:name", controllers.GetDocTypeMeta)
		api.GET("/resource/:doctype", controllers.ListResources)
		api.GET("/resource/:doctype/single", controllers.GetSingleResource)
		api.PUT("/resource/:doctype/single", controllers.UpdateSingleResource)
		api.DELETE("/resource/:doctype/single", controllers.DeleteSingleResource)
		api.GET("/resource/:doctype/link-search", controllers.SearchLinkOptions)
		api.POST("/resource/:doctype", controllers.CreateResource)
		api.GET("/resource/:doctype/:name", controllers.GetResource)
		api.PUT("/resource/:doctype/:name", controllers.UpdateResource)
		api.DELETE("/resource/:doctype/:name", controllers.DeleteResource)
		api.GET("/files", controllers.ListFiles)
		api.POST("/files/upload", controllers.UploadFile)
	}

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
