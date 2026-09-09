package repository

import (
	"errors"
	"judge-backend/internal/app/contest_domain/activity/domain/entity/clarification"
)

var (
	ErrorContestProjectNotFound = errors.New("contest project not found")
	ErrorReceiverUserNotFound   = errors.New("receiver user not found")
)

type ClarificationRepository interface {
	FindByID(id int32) (*clarification.Clarification, error)
	Save(clarification *clarification.Clarification) error
	Delete(clarification *clarification.Clarification) error
}
