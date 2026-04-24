package metadata

import (
	"strings"

	"gorm.io/gorm"
)

func LoadDocTypeByName(db *gorm.DB, name string, withFields bool) (*DocType, error) {
	var dt DocType
	q := db.Model(&DocType{})
	if withFields {
		q = q.Preload("Fields", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("sort_order ASC, id ASC")
		})
	}
	if err := q.Where("LOWER(name) = LOWER(?)", strings.TrimSpace(name)).First(&dt).Error; err != nil {
		return nil, err
	}
	return &dt, nil
}
