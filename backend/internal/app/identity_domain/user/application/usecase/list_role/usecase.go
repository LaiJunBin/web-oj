package listrole

import (
	"errors"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/identity_domain/user/application/service/query/rolequery"
)

var (
	ErrorInternalServer = errors.New("internal server error")
)

type Input struct{}

type Output struct {
	Roles []rolequery.Role
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	roleQueryService rolequery.RoleQueryService
	logger           sharedtypes.Logger
}

func NewUsecase(
	roleQueryService rolequery.RoleQueryService,
	logger sharedtypes.Logger,
) Usecase {
	return &usecase{
		roleQueryService: roleQueryService,
		logger:           logger,
	}
}

func (uc *usecase) Execute(input *Input) (*Output, error) {
	uc.logger.Infof("Starting to list roles")

	roles, err := uc.roleQueryService.ListRole()
	if err != nil {
		uc.logger.Errorf("Failed to list roles", "error", err)
		return nil, ErrorInternalServer
	}

	uc.logger.Infof("Successfully listed roles", "count", len(roles))
	return &Output{
		Roles: roles,
	}, nil
}
