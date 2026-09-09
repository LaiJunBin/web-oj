package repository

import (
	"judge-backend/internal/app/contest_domain/contest/domain/entity/clarification"
)

type ClarificationRepository interface {
	FindByID(id int32) (*clarification.Clarification, error)
	Save(clarification *clarification.Clarification) error
}
