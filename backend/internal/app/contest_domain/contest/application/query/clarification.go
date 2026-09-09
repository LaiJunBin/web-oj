package query

import "judge-backend/internal/app/contest_domain/contest/application/query/dto"

type ClarificationQuery interface {
	ListUserClarificationsByProjectID(userID, contestProjectID int32) ([]*dto.Clarification, error)
	GetUserClarificationByID(userID, clarificationID int32) (*dto.ClarificationDetail, error)
}
