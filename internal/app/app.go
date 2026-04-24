package app

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"gogal/internal/config"
	"gogal/internal/crud"
	"gogal/internal/database"
	"gogal/internal/studio"

	"github.com/gin-gonic/gin"
)

func RunServer(ctx context.Context) error {
	runtimeCfg, err := config.LoadRuntimeConfig()
	if err != nil {
		return err
	}
	_ = os.Setenv("DB_HOST", runtimeCfg.Site.DBHost)
	_ = os.Setenv("DB_PORT", fmt.Sprintf("%d", runtimeCfg.Site.DBPort))
	_ = os.Setenv("DB_NAME", runtimeCfg.Site.DBName)
	_ = os.Setenv("DB_USER", runtimeCfg.Site.DBUser)
	_ = os.Setenv("DB_PASSWORD", runtimeCfg.Site.DBPassword)
	dbCfg := database.DBConfig{
		Host:     runtimeCfg.Site.DBHost,
		Port:     runtimeCfg.Site.DBPort,
		Name:     runtimeCfg.Site.DBName,
		User:     runtimeCfg.Site.DBUser,
		Password: runtimeCfg.Site.DBPassword,
	}
	if _, err := database.Connect(dbCfg); err != nil {
		return err
	}
	if err := database.RunMetadataMigration(); err != nil {
		return err
	}

	r := gin.Default()
	tmpl := template.Must(template.ParseGlob("views/*.html"))
	tmpl = template.Must(tmpl.ParseGlob("views/partials/*.html"))
	r.SetHTMLTemplate(tmpl)
	r.Static("/public", "./public")
	r.Static("/files", "./storage/public")

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Welcome to Gogal API", "status": "active"})
	})

	api := r.Group("/api")
	crud.RegisterAPIRoutes(api)
	studio.RegisterRoutes(r)

	srv := &http.Server{Addr: fmt.Sprintf(":%d", runtimeCfg.HTTPPort), Handler: r}
	errCh := make(chan error, 1)
	go func() {
		log.Printf("Gogal server running on :%d", runtimeCfg.HTTPPort)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	case err := <-errCh:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}
}
