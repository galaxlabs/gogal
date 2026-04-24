package database

import (
	"fmt"
	"log"

	legacyconfig "gogal/config"
	"gogal/internal/migration"
)

func RunMetadataMigration() error {
	if legacyconfig.DB == nil {
		return fmt.Errorf("database is not connected")
	}
	exec := migration.NewExecutor(legacyconfig.DB)
	if err := exec.SyncAllDocTypes(); err != nil {
		return err
	}
	log.Println("metadata migration completed")
	return nil
}
