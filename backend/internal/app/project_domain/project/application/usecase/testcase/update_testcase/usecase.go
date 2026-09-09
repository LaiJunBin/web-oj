package updatetestcase

import (
	"errors"
	"judge-backend/internal/app/project_domain/project/domain/repository"
)

var (
	ErrorInvalidID              = errors.New("invalid id, must be greater than 0")
	ErrorEmptyInput             = errors.New("empty input")
	ErrorInvalidName            = errors.New("invalid name")
	ErrorInvalidScore           = errors.New("invalid score")
	ErrorTestcaseNotFound       = errors.New("testcase not found")
	ErrorDependTestcaseNotFound = errors.New("depend testcase not found")
	ErrorInternalServer         = errors.New("internal server error")
)

type Input struct {
	ID       int32
	Name     *string
	Command  *string
	Score    *float64
	IsSample *bool
	Depends  *[]int32
}

type Output struct {
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	repository repository.TestcaseRepository
}

func NewUsecase(
	repository repository.TestcaseRepository,
) Usecase {
	return &usecase{
		repository: repository,
	}
}

func (u *usecase) validate(input *Input) error {
	if input.ID <= 0 {
		return ErrorInvalidID
	}

	if input.Name == nil && input.Command == nil && input.Score == nil && input.IsSample == nil && input.Depends == nil {
		return ErrorEmptyInput
	}

	if input.Score != nil && *input.Score < 0 {
		return ErrorInvalidScore
	}

	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	if err := u.validate(input); err != nil {
		return nil, err
	}

	tc, err := u.repository.FindByID(int(input.ID))
	if err != nil {
		return nil, ErrorInternalServer
	}

	if tc == nil {
		return nil, ErrorTestcaseNotFound
	}

	if input.Name != nil {
		if err := tc.SetName(*input.Name); err != nil {
			return nil, ErrorInvalidName
		}
	}

	if input.Command != nil {
		tc.SetCommand(*input.Command)
	}

	if input.Score != nil {
		tc.SetScore(*input.Score)
	}

	if input.IsSample != nil {
		tc.SetIsSample(*input.IsSample)
	}

	if input.Depends != nil {
		if err := tc.SetDepends(*input.Depends); err != nil {
			return nil, err
		}
	}

	if err := u.repository.Save(tc); err != nil {
		if err == repository.ErrorDependTestcaseNotFound {
			return nil, ErrorDependTestcaseNotFound
		}
		return nil, ErrorInternalServer
	}

	return &Output{}, nil
}
