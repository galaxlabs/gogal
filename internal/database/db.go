package database

import (
	"fmt"
	"log"
	"time"

	legacyconfig "gogal/config"

	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect(cfg DBConfig) (*gorm.DB, error) {
	db, err := legacyconfig.OpenWithDSN(cfg.DSN())
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("sql db handle: %w", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("database ping: %w", err)
	}
	DB = db
	legacyconfig.DB = db
	log.Printf("database connected: host=%s port=%d db=%s user=%s", cfg.Host, cfg.Port, cfg.Name, cfg.User)
	return db, nil
}
