package repository

import (
	"errors"
	"judge-backend/internal/app/project_domain/project/domain/entity/gradingitem"
)

var (
	ErrorDependTestcaseNotFound          = errors.New("depend testcase not found")
	ErrorDependTestcaseIsSameGradingItem = errors.New("depend testcase is in the same grading item")
)

type GradingItemRepository interface {
	Exists(id int) (bool, error)
	FindByID(id int) (*gradingitem.GradingItem, error)
	Save(gradingItem *gradingitem.GradingItem) error
	Delete(gradingItem *gradingitem.GradingItem) error
}
