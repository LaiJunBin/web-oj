package creategradingitem

import (
	"errors"
	"judge-backend/internal/app/project_domain/project/domain/entity/gradingitem"
	"judge-backend/internal/app/project_domain/project/domain/entity/project"
	"judge-backend/internal/app/project_domain/project/domain/repository"
	"judge-backend/internal/app/project_domain/project/domain/service"
)

var (
	ErrorInvalidProjectID       = errors.New("invalid project id, must be greater than 0")
	ErrorEmptyName              = errors.New("empty name")
	ErrorEmptyLabel             = errors.New("empty label")
	ErrorInvalidScore           = errors.New("invalid score, must be greater than or equal to 0")
	ErrorEmptyScoringStrategy   = errors.New("empty scoring strategy")
	ErrorEmptyOutputFormat      = errors.New("empty output format")
	ErrorInvalidResultParser    = errors.New("invalid result parser")
	ErrorProjectNotFound        = errors.New("project not found")
	ErrorDependTestcaseNotFound = errors.New("depend testcase not found")
	ErrorInternalServer         = errors.New("internal server error")
)

type Input struct {
	ProjectID       int32
	Name            string
	Description     string
	Score           float64
	ScoringStrategy string
	OutputFormat    string
	ResultParser    string
	Label           string
	DependTestcases []int32
}

type Output struct {
	ID int32
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	repository             repository.GradingItemRepository
	projectRepo            repository.ProjectRepository
	resultParserValidator  service.ResultParserValidator
}

func NewUsecase(
	repository repository.GradingItemRepository,
	projectRepo repository.ProjectRepository,
	resultParserValidator service.ResultParserValidator,
) Usecase {
	return &usecase{
		repository:            repository,
		projectRepo:           projectRepo,
		resultParserValidator: resultParserValidator,
	}
}

func (u *usecase) validate(input *Input) error {
	if input.ProjectID <= 0 {
		return ErrorInvalidProjectID
	}

	if input.Name == "" {
		return ErrorEmptyName
	}

	if input.Label == "" {
		return ErrorEmptyLabel
	}

	if input.Score < 0 {
		return ErrorInvalidScore
	}

	if input.ScoringStrategy == "" {
		return ErrorEmptyScoringStrategy
	}

	if input.OutputFormat == "" {
		return ErrorEmptyOutputFormat
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

	if exists, err := u.projectRepo.Exists(int(input.ProjectID)); err != nil {
		return nil, ErrorInternalServer
	} else if !exists {
		return nil, ErrorProjectNotFound
	}

	gradingItem, err := gradingitem.New(&gradingitem.GradingItemConfig{
		ProjectID:       project.ID(input.ProjectID),
		Name:            input.Name,
		Description:     input.Description,
		Score:           input.Score,
		ScoringStrategy: input.ScoringStrategy,
		OutputFormat:    input.OutputFormat,
		ResultParser:    input.ResultParser,
		Label:           input.Label,
		DependTestcases: input.DependTestcases,
	})
	if err != nil {
		return nil, err
	}

	if err := u.repository.Save(gradingItem); err != nil {
		if errors.Is(err, repository.ErrorDependTestcaseNotFound) {
			return nil, ErrorDependTestcaseNotFound
		}

		return nil, ErrorInternalServer
	}

	return &Output{
		ID: gradingItem.ID(),
	}, nil
}
