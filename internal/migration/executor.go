package migration

import (
	"fmt"

	"gorm.io/gorm"
)

type Executor struct {
	db *gorm.DB
}

func NewExecutor(db *gorm.DB) *Executor {
	return &Executor{db: db}
}

func (e *Executor) BuildPlan() (*Plan, error) {
	return BuildPlan(e.db)
}

func (e *Executor) ApplyPlan(plan *Plan) error {
	if plan == nil {
		return nil
	}
	for _, item := range plan.Items {
		if err := e.db.Exec(item.Statement).Error; err != nil {
			return fmt.Errorf("apply %s for %s (%s): %w", item.Type, item.DocType, item.Statement, err)
		}
	}
	return nil
}

func (e *Executor) SyncAllDocTypes() error {
	plan, err := e.BuildPlan()
	if err != nil {
		return err
	}
	return e.ApplyPlan(plan)
}
