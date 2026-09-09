package getuser

import (
	"errors"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/identity_domain/user/application/service/query/userquery"
)

var (
	ErrorInvalidID      = errors.New("invalid ID")
	ErrorUserNotFound   = errors.New("user not found")
	ErrorInternalServer = errors.New("internal server error")
)

type Input struct {
	ID int32
}

type Output struct {
	User *userquery.User
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	userQueryService userquery.UserQueryService
	logger           sharedtypes.Logger
}

func NewUsecase(
	userQueryService userquery.UserQueryService,
	logger sharedtypes.Logger,
) Usecase {
	return &usecase{
		userQueryService: userQueryService,
		logger:           logger,
	}
}

func (u *usecase) validate(input *Input) error {
	if input.ID <= 0 {
		return ErrorInvalidID
	}

	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	u.logger.Infof("Starting to get user", "userId", input.ID)

	if err := u.validate(input); err != nil {
		u.logger.Errorf("Failed to validate input", "userId", input.ID, "error", err)
		return nil, err
	}

	user, err := u.userQueryService.GetUserByID(input.ID)
	if err != nil {
		switch {
		case errors.Is(err, userquery.ErrorUserNotFound):
			u.logger.Infof("User not found", "userId", input.ID)
			return nil, ErrorUserNotFound
		default:
			u.logger.Errorf("Failed to get user", "userId", input.ID, "error", err)
			return nil, ErrorInternalServer
		}
	}

	u.logger.Infof("Successfully got user", "userId", input.ID)
	return &Output{
		User: user,
	}, nil
}
