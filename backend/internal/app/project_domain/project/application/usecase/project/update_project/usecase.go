package updateproject

import (
	"errors"
	"judge-backend/internal/app/project_domain/project/domain/entity/project"
	"judge-backend/internal/app/project_domain/project/domain/repository"
	"judge-backend/internal/app/project_domain/project/domain/service"
)

var (
	ErrorNoInput                = errors.New("no input provided")
	ErrorNameIsRequired         = errors.New("name is required")
	ErrorImageTagIsRequired     = errors.New("image tag is required")
	ErrorTestImageTagIsRequired = errors.New("test image tag is required")
	ErrorProjectNotFound        = errors.New("project not found")
	ErrorImageNotFound          = errors.New("project image not found")
	ErrorTestImageNotFound      = errors.New("test image not found")
	ErrorInvalidResultParser    = errors.New("invalid result parser")
	ErrorInternalServer         = errors.New("internal server error")
)

type Input struct {
	ID                  int32
	Name                *string
	Description         *string
	ImageTag            *string
	TestImageTag        *string
	Env                 *string
	TestEnv             *string
	ResultParser        *string
	DynamicFieldConfigs *[]*project.DynamicFieldConfig
}

type Output struct {
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
	if input.Name == nil &&
		input.Description == nil &&
		input.ImageTag == nil &&
		input.TestImageTag == nil &&
		input.Env == nil &&
		input.TestEnv == nil &&
		input.ResultParser == nil &&
		input.DynamicFieldConfigs == nil {
		return ErrorNoInput
	}

	if input.Name != nil && *input.Name == "" {
		return ErrorNameIsRequired
	}

	if input.ImageTag != nil && *input.ImageTag == "" {
		return ErrorImageTagIsRequired
	}

	if input.TestImageTag != nil && *input.TestImageTag == "" {
		return ErrorTestImageTagIsRequired
	}

	if input.ResultParser != nil && *input.ResultParser != "" && !u.resultParserValidator.IsValid(*input.ResultParser) {
		return ErrorInvalidResultParser
	}

	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	if err := u.validate(input); err != nil {
		return nil, err
	}

	project, err := u.repository.FindByID(int(input.ID))
	if err != nil {
		return nil, ErrorInternalServer
	}

	if project == nil {
		return nil, ErrorProjectNotFound
	}

	if input.ImageTag != nil {
		if exists, err := u.imageRepo.Exists(*input.ImageTag); err != nil {
			return nil, ErrorInternalServer
		} else if !exists {
			return nil, ErrorImageNotFound
		}

		project.SetImage(*input.ImageTag)
	}

	if input.TestImageTag != nil {
		if exists, err := u.imageRepo.Exists(*input.TestImageTag); err != nil {
			return nil, ErrorInternalServer
		} else if !exists {
			return nil, ErrorTestImageNotFound
		}

		project.SetTestImage(*input.TestImageTag)
	}

	if input.Name != nil {
		if err := project.SetName(*input.Name); err != nil {
			return nil, err
		}
	}

	if input.Description != nil {
		project.SetDescription(*input.Description)
	}

	if input.Env != nil {
		if err := project.SetEnv(*input.Env); err != nil {
			return nil, err
		}
	}

	if input.TestEnv != nil {
		if err := project.SetTestEnv(*input.TestEnv); err != nil {
			return nil, err
		}
	}

	if input.ResultParser != nil {
		project.SetResultParser(*input.ResultParser)
	}

	if input.DynamicFieldConfigs != nil {
		if err := project.SetDynamicFields(*input.DynamicFieldConfigs); err != nil {
			return nil, err
		}
	}

	if err := u.repository.Save(project); err != nil {
		return nil, ErrorInternalServer
	}

	return &Output{}, nil
}
