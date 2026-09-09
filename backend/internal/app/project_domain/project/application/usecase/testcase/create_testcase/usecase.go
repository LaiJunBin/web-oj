package createtestcase

import (
	"errors"
	"judge-backend/internal/app/project_domain/project/domain/entity/testcase"
	"judge-backend/internal/app/project_domain/project/domain/repository"
)

var (
	ErrorInvalidGradingItemID   = errors.New("invalid grading item id")
	ErrorEmptyName              = errors.New("empty name")
	ErrorEmptyCommand           = errors.New("empty command")
	ErrorInvalidScore           = errors.New("invalid score")
	ErrorDependTestcaseNotFound = errors.New("depend testcase not found")
	ErrorGradingItemNotFound    = errors.New("grading item not found")
	ErrorInternalServer         = errors.New("internal server error")
)

type Input struct {
	GradingItemID int32
	Name          string
	Command       string
	Score         float64
	IsSample      bool
	Depends       []int32
}

type Output struct {
	ID int32
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	repository      repository.TestcaseRepository
	gradingItemRepo repository.GradingItemRepository
}

func NewUsecase(
	repository repository.TestcaseRepository,
	gradingItemRepo repository.GradingItemRepository,
) Usecase {
	return &usecase{
		repository:      repository,
		gradingItemRepo: gradingItemRepo,
	}
}

func (u *usecase) validate(input *Input) error {
	if input.GradingItemID <= 0 {
		return ErrorInvalidGradingItemID
	}

	if input.Name == "" {
		return ErrorEmptyName
	}

	if input.Command == "" {
		return ErrorEmptyCommand
	}

	if input.Score < 0 {
		return ErrorInvalidScore
	}

	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	if err := u.validate(input); err != nil {
		return nil, err
	}

	if exists, err := u.gradingItemRepo.Exists(int(input.GradingItemID)); err != nil {
		return nil, ErrorInternalServer
	} else if !exists {
		return nil, ErrorGradingItemNotFound
	}

	tc, err := testcase.New(&testcase.TestcaseConfig{
		GradingItemID: input.GradingItemID,
		Name:          input.Name,
		Command:       input.Command,
		Score:         input.Score,
		IsSample:      input.IsSample,
		Depends:       input.Depends,
	})
	if err != nil {
		return nil, err
	}

	if err := u.repository.Save(tc); err != nil {
		if err == repository.ErrorDependTestcaseNotFound {
			return nil, ErrorDependTestcaseNotFound
		}
		return nil, ErrorInternalServer
	}

	return &Output{
		ID: tc.ID(),
	}, nil
}
