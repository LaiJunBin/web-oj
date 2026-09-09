package query

import "judge-backend/internal/app/contest_domain/activity/application/query/dto"

type ClarificationQuery interface {
	ListClarifications(contestProjectID int32) ([]*dto.Clarification, error)
	GetClarificationByID(id int32) (*dto.ClarificationDetail, error)
}
