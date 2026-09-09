package listuser

import (
	"errors"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/identity_domain/user/application/service/query/userquery"
)

var (
	ErrorInternalServer = errors.New("internal server error")
)

type Input struct {
}

type Output struct {
	Users []userquery.User
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	service userquery.UserQueryService
	logger  sharedtypes.Logger
}

func NewUsecase(
	service userquery.UserQueryService,
	logger sharedtypes.Logger,
) Usecase {
	return &usecase{
		service: service,
		logger:  logger,
	}
}

func (u *usecase) validate(input *Input) error {
	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	u.logger.Infof("Starting to list users")

	if err := u.validate(input); err != nil {
		u.logger.Errorf("Failed to validate input", "error", err)
		return nil, err
	}

	users, err := u.service.ListUser()
	if err != nil {
		u.logger.Errorf("Failed to list users", "error", err)
		return nil, ErrorInternalServer
	}

	u.logger.Infof("Successfully listed users", "count", len(users))
	return &Output{
		Users: users,
	}, nil
}
