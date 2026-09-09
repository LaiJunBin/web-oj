package getsessionttl

import (
	"judge-backend/internal/app/contest_domain/contest/application/query"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/shared/types/httperror"
	"net/http"
)

var (
	ErrorInvalidSessionID = httperror.New(http.StatusBadRequest, "invalid session id")
)

type Input struct {
	SessionID string
}

type Output struct {
	TTL int
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	query  query.SessionQuery
	logger sharedtypes.Logger
}

func NewUsecase(
	query query.SessionQuery,
	logger sharedtypes.Logger,
) Usecase {
	return &usecase{
		query:  query,
		logger: logger,
	}
}

func (u *usecase) validate(input *Input) error {
	if input.SessionID == "" {
		return ErrorInvalidSessionID
	}

	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	u.logger.Infof("Starting to get session TTL", "sessionId", input.SessionID)

	if err := u.validate(input); err != nil {
		u.logger.Errorf("Failed to validate input", "sessionId", input.SessionID, "error", err)
		return nil, err
	}

	ttl := u.query.GetTTL(input.SessionID)
	u.logger.Infof("Successfully got session TTL", "sessionId", input.SessionID, "ttl", ttl)

	return &Output{
		TTL: ttl,
	}, nil
}
