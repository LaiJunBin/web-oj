package createlifecycle

import (
	"errors"
	"judge-backend/internal/app/project_domain/project/domain/entity/lifecycle"
	"judge-backend/internal/app/project_domain/project/domain/entity/project"
	"judge-backend/internal/app/project_domain/project/domain/repository"
)

var (
	ErrorInvalidProjectID            = errors.New("invalid project id, must be greater than 0")
	ErrorEmptyType                   = errors.New("type is required")
	ErrorEmptyName                   = errors.New("name is required")
	ErrorEmptyCommand                = errors.New("command is required")
	ErrorEmptyExecutionEnvironment   = errors.New("execution environment is required")
	ErrorInternalServer              = errors.New("internal server error")
	ErrorProjectNotFound             = errors.New("project not found")
	ErrorInvalidType                 = errors.New("invalid type, must be one of before_all, before_each, after_each, after_all")
	ErrorInvalidExecutionEnvironment = errors.New("invalid execution environment, must be one of project, test")
)

type Input struct {
	ProjectID            int32
	Type                 string
	Name                 string
	Description          string
	Command              string
	ExecutionEnvironment string
}

type Output struct {
	ID int32
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	repository  repository.LifecycleRepository
	projectRepo repository.ProjectRepository
}

func NewUsecase(
	repository repository.LifecycleRepository,
	projectRepo repository.ProjectRepository,
) Usecase {
	return &usecase{
		repository:  repository,
		projectRepo: projectRepo,
	}
}

func (u *usecase) validate(input *Input) error {
	if input.ProjectID <= 0 {
		return ErrorInvalidProjectID
	}

	if input.Type == "" {
		return ErrorEmptyType
	}

	if input.Name == "" {
		return ErrorEmptyName
	}

	if input.Command == "" {
		return ErrorEmptyCommand
	}

	if input.ExecutionEnvironment == "" {
		return ErrorEmptyExecutionEnvironment
	}

	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	if err := u.validate(input); err != nil {
		return nil, err
	}

	if exists, err := u.projectRepo.Exists(int(input.ProjectID)); err != nil {
		return nil, ErrorInternalServer
	} else if !exists {
		return nil, ErrorProjectNotFound
	}

	if _, ok := lifecycle.Types[input.Type]; !ok {
		return nil, ErrorInvalidType
	}

	execEnv, ok := lifecycle.ExecutionEnvironments[input.ExecutionEnvironment]
	if !ok {
		return nil, ErrorInvalidExecutionEnvironment
	}

	lc, err := lifecycle.New(&lifecycle.LifecycleConfig{
		ProjectID:            project.ID(input.ProjectID),
		ExecutionEnvironment: execEnv.String(),
		Type:                 input.Type,
		Name:                 input.Name,
		Description:          input.Description,
		Command:              input.Command,
	})
	if err != nil {
		return nil, err
	}

	if err := u.repository.Save(lc); err != nil {
		return nil, ErrorInternalServer
	}

	return &Output{
		ID: lc.ID().Value(),
	}, nil
}
