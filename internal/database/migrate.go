package database

import (
	"fmt"
	"log"

	legacyconfig "gogal/config"
	"gogal/internal/migration"
	"gogal/models"
)

func RunMetadataMigration() error {
	if legacyconfig.DB == nil {
		return fmt.Errorf("database is not connected")
	}
	if err := legacyconfig.DB.AutoMigrate(&models.AuditLog{}, &models.Comment{}, &models.Assignment{}); err != nil {
		return fmt.Errorf("migrate activity tables: %w", err)
	}
	exec := migration.NewExecutor(legacyconfig.DB)
	if err := exec.SyncAllDocTypes(); err != nil {
		return err
	}
	log.Println("metadata migration completed")
	return nil
}
