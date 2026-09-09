package repository

import (
	"errors"

	"judge-backend/internal/app/contest_domain/activity/domain/entity/judgment"
)

// ErrJudgmentNotFound is what FindByID returns for an unknown id, so a
// usecase can decide the HTTP status without knowing the storage driver.
var ErrJudgmentNotFound = errors.New("judgment not found")

type JudgmentRepository interface {
	FindByID(id string) (*judgment.Judgment, error)
}
