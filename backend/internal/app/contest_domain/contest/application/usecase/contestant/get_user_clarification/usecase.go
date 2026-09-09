package getuserclarification

import (
	"judge-backend/internal/app/contest_domain/contest/application/query"
	"judge-backend/internal/app/contest_domain/contest/application/query/dto"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/shared/types/httperror"
	"net/http"
)

var (
	ErrorClarificationNotFound = httperror.New(http.StatusNotFound, "clarification not found")
)

type Input struct {
	UserID          int32
	ClarificationID int32
}

type Output struct {
	Clarification *dto.ClarificationDetail
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	query  query.ClarificationQuery
	logger sharedtypes.Logger
}

func NewUsecase(
	query query.ClarificationQuery,
	logger sharedtypes.Logger,
) Usecase {
	return &usecase{
		query:  query,
		logger: logger,
	}
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	u.logger.Infof("Starting to get user clarification", "userId", input.UserID, "clarificationId", input.ClarificationID)

	clarification, err := u.query.GetUserClarificationByID(input.UserID, input.ClarificationID)
	if err != nil {
		u.logger.Errorf("Failed to get user clarification from database", "userId", input.UserID, "clarificationId", input.ClarificationID, "error", err)
		return nil, err
	}

	if clarification == nil {
		u.logger.Errorf("Clarification not found", "userId", input.UserID, "clarificationId", input.ClarificationID)
		return nil, ErrorClarificationNotFound
	}

	u.logger.Infof("Successfully got user clarification", "userId", input.UserID, "clarificationId", input.ClarificationID)
	return &Output{
		Clarification: clarification,
	}, nil
}
