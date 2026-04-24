package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"gogal/models"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDB() error {
	_ = godotenv.Load()

	db, err := OpenWithDSN(buildDSN())
	if err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("access sql db handle: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	if err := db.AutoMigrate(&models.DocType{}, &models.DocField{}, &models.File{}, &models.SingleValue{}, &models.AuditLog{}, &models.Comment{}, &models.Assignment{}); err != nil {
		return fmt.Errorf("auto migrate metadata tables: %w", err)
	}

	if err := seedSystemDocTypes(db); err != nil {
		return fmt.Errorf("seed system doctypes: %w", err)
	}

	DB = db
	log.Println("database connected and metadata tables are ready")
	return nil
}

func OpenWithDSN(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}
	return db, nil
}

func buildDSN() string {
	values := map[string]string{
		"host":     getEnv("DB_HOST", "localhost"),
		"user":     getEnv("DB_USER", "gogaluser"),
		"password": getEnv("DB_PASSWORD", "gogal123"),
		"dbname":   getEnv("DB_NAME", "gogaldb"),
		"port":     getEnv("DB_PORT", "5432"),
		"sslmode":  getEnv("DB_SSLMODE", "disable"),
		"TimeZone": getEnv("DB_TIMEZONE", "UTC"),
	}

	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		values["host"],
		values["user"],
		values["password"],
		values["dbname"],
		values["port"],
		values["sslmode"],
		values["TimeZone"],
	)
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func seedSystemDocTypes(db *gorm.DB) error {
	for _, systemDocType := range models.SystemDocTypes() {
		if err := systemDocType.Normalize(); err != nil {
			return err
		}

		var existing models.DocType
		err := db.Preload("Fields", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("sort_order ASC, id ASC")
		}).Where("name = ?", systemDocType.Name).First(&existing).Error

		switch {
		case err == nil:
			if err := syncSystemDocFields(db, existing.ID, systemDocType.Fields); err != nil {
				return err
			}
		case errors.Is(err, gorm.ErrRecordNotFound):
			for index := range systemDocType.Fields {
				systemDocType.Fields[index].SortOrder = index + 1
			}
			if err := db.Create(&systemDocType).Error; err != nil {
				return err
			}
		default:
			return err
		}
	}

	return nil
}

func syncSystemDocFields(db *gorm.DB, docTypeID uint, fields []models.DocField) error {
	for index := range fields {
		field := fields[index]
		field.DocTypeID = docTypeID
		field.SortOrder = index + 1

		var existing models.DocField
		err := db.Where("doc_type_id = ? AND field_name = ?", docTypeID, field.FieldName).First(&existing).Error
		switch {
		case err == nil:
			existing.Label = field.Label
			existing.FieldType = field.FieldType
			existing.Options = field.Options
			existing.DefaultValue = field.DefaultValue
			existing.Required = field.Required
			existing.ReadOnly = field.ReadOnly
			existing.Hidden = field.Hidden
			existing.Unique = field.Unique
			existing.InListView = field.InListView
			existing.SortOrder = field.SortOrder
			if err := db.Save(&existing).Error; err != nil {
				return err
			}
		case errors.Is(err, gorm.ErrRecordNotFound):
			if err := db.Create(&field).Error; err != nil {
				return err
			}
		default:
			return err
		}
	}

	return nil
}
