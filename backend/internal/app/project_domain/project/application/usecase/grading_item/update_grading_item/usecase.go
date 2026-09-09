package updategradingitem

import (
	"errors"
	"judge-backend/internal/app/project_domain/project/domain/entity/gradingitem/outputformater"
	"judge-backend/internal/app/project_domain/project/domain/entity/gradingitem/scoringstrategy"
	"judge-backend/internal/app/project_domain/project/domain/repository"
	"judge-backend/internal/app/project_domain/project/domain/service"
)

var (
	ErrorInvalidID                       = errors.New("invalid id, must be greater than 0")
	ErrorEmptyInput                      = errors.New("empty input")
	ErrorGradingItemNotFound             = errors.New("grading item not found")
	ErrorInvalidName                     = errors.New("invalid name")
	ErrorEmptyLabel                      = errors.New("empty label")
	ErrorInvalidScore                    = errors.New("invalid score")
	ErrorInvalidScoringStrategy          = errors.New("invalid scoring strategy")
	ErrorInvalidOutputFormat             = errors.New("invalid output format")
	ErrorInvalidResultParser             = errors.New("invalid result parser")
	ErrorDependTestcaseNotFound          = errors.New("depend testcase not found")
	ErrorDependTestcaseIsSameGradingItem = errors.New("depend testcase is same grading item")
	ErrorInternalServer                  = errors.New("internal server error")
)

type Input struct {
	ID              int32
	Name            *string
	Description     *string
	Score           *float64
	ScoringStrategy *string
	OutputFormat    *string
	ResultParser    *string
	Label           *string
	DependTestcases *[]int32
}

type Output struct {
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	repository            repository.GradingItemRepository
	resultParserValidator service.ResultParserValidator
}

func NewUsecase(
	repository repository.GradingItemRepository,
	resultParserValidator service.ResultParserValidator,
) Usecase {
	return &usecase{
		repository:            repository,
		resultParserValidator: resultParserValidator,
	}
}

func (u *usecase) validate(input *Input) error {
	if input.ID <= 0 {
		return ErrorInvalidID
	}

	if input.Name == nil && input.Description == nil && input.Score == nil && input.ScoringStrategy == nil && input.OutputFormat == nil && input.ResultParser == nil && input.Label == nil && input.DependTestcases == nil {
		return ErrorEmptyInput
	}

	if input.Score != nil && *input.Score < 0 {
		return ErrorInvalidScore
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

	gradingItem, err := u.repository.FindByID(int(input.ID))
	if err != nil {
		return nil, ErrorInternalServer
	}

	if gradingItem == nil {
		return nil, ErrorGradingItemNotFound
	}

	if input.Name != nil {
		if err := gradingItem.SetName(*input.Name); err != nil {
			return nil, ErrorInvalidName
		}
	}

	if input.Description != nil {
		gradingItem.SetDescription(*input.Description)
	}

	if input.Score != nil {
		gradingItem.SetScore(*input.Score)
	}

	if input.ScoringStrategy != nil {
		st, ok := scoringstrategy.Strategies[*input.ScoringStrategy]
		if !ok {
			return nil, ErrorInvalidScoringStrategy
		}
		gradingItem.SetScoringStrategy(st)
	}

	if input.OutputFormat != nil {
		f, ok := outputformater.Formaters[*input.OutputFormat]
		if !ok {
			return nil, ErrorInvalidOutputFormat
		}
		gradingItem.SetOutputFormater(f)
	}

	if input.ResultParser != nil {
		gradingItem.SetResultParser(*input.ResultParser)
	}

	if input.Label != nil {
		if *input.Label == "" {
			return nil, ErrorEmptyLabel
		}
		gradingItem.SetLabel(*input.Label)
	}

	if input.DependTestcases != nil {
		gradingItem.SetDependTestcases(*input.DependTestcases)
	}

	if err := u.repository.Save(gradingItem); err != nil {
		if err == repository.ErrorDependTestcaseNotFound {
			return nil, ErrorDependTestcaseNotFound
		} else if err == repository.ErrorDependTestcaseIsSameGradingItem {
			return nil, ErrorDependTestcaseIsSameGradingItem
		}
		return nil, ErrorInternalServer
	}

	return &Output{}, nil
}
