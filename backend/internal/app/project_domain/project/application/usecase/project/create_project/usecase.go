package createproject

import (
	"errors"
	"judge-backend/internal/app/project_domain/project/domain/entity/project"
	"judge-backend/internal/app/project_domain/project/domain/repository"
	"judge-backend/internal/app/project_domain/project/domain/service"
)

var (
	ErrorNameIsRequired         = errors.New("name is required")
	ErrorImageTagIsRequired     = errors.New("image tag is required")
	ErrorTestImageTagIsRequired = errors.New("test image tag is required")
	ErrorImageNotFound          = errors.New("project image not found")
	ErrorTestImageNotFound      = errors.New("test image not found")
	ErrorInvalidResultParser    = errors.New("invalid result parser")
	ErrorInternalServer         = errors.New("internal server error")
)

type Input struct {
	Name                string
	Description         string
	ImageTag            string
	TestImageTag        string
	Env                 string
	TestEnv             string
	ResultParser        string
	DynamicFieldConfigs []*project.DynamicFieldConfig
}

type Output struct {
	ID int32
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	repository            repository.ProjectRepository
	imageRepo             repository.ImageRepository
	resultParserValidator service.ResultParserValidator
}

func NewUsecase(
	repository repository.ProjectRepository,
	imageRepo repository.ImageRepository,
	resultParserValidator service.ResultParserValidator,
) Usecase {
	return &usecase{
		repository:            repository,
		imageRepo:             imageRepo,
		resultParserValidator: resultParserValidator,
	}
}

func (u *usecase) validate(input *Input) error {
	if input.Name == "" {
		return ErrorNameIsRequired
	}

	if input.ImageTag == "" {
		return ErrorImageTagIsRequired
	}

	if input.TestImageTag == "" {
		return ErrorTestImageTagIsRequired
	}

	if input.ResultParser != "" && !u.resultParserValidator.IsValid(input.ResultParser) {
		return ErrorInvalidResultParser
	}

	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	if err := u.validate(input); err != nil {
		return nil, err
	}

	if exists, err := u.imageRepo.Exists(input.ImageTag); err != nil {
		return nil, ErrorInternalServer
	} else if !exists {
		return nil, ErrorImageNotFound
	}

	if exists, err := u.imageRepo.Exists(input.TestImageTag); err != nil {
		return nil, ErrorInternalServer
	} else if !exists {
		return nil, ErrorTestImageNotFound
	}

	project, err := project.New(&project.ProjectConfig{
		Name:          input.Name,
		Description:   input.Description,
		Image:         &input.ImageTag,
		TestImage:     &input.TestImageTag,
		Env:           input.Env,
		TestEnv:       input.TestEnv,
		ResultParser:  input.ResultParser,
		DynamicFields: input.DynamicFieldConfigs,
	})
	if err != nil {
		return nil, err
	}

	if err := u.repository.Save(project); err != nil {
		return nil, ErrorInternalServer
	}

	return &Output{
		ID: project.ID().Value(),
	}, nil
}
