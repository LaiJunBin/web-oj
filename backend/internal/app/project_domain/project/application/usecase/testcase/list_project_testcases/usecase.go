package listprojecttestcases

import (
	"errors"
	"judge-backend/internal/app/project_domain/project/application/query"
	"judge-backend/internal/app/project_domain/project/application/query/dto"
)

var (
	ErrorInvalidProjectID = errors.New("invalid project id, must be greater than 0")
	ErrorInternalServer   = errors.New("internal server error")
)

type Input struct {
	ProjectID int32
}

type Output struct {
	GradingItems []*dto.ProjectTestcaseGradingItem
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	query query.TestcaseQuery
}

func NewUsecase(
	query query.TestcaseQuery,
) Usecase {
	return &usecase{
		query: query,
	}
}

func (u *usecase) validate(input *Input) error {
	if input.ProjectID <= 0 {
		return ErrorInvalidProjectID
	}

	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	if err := u.validate(input); err != nil {
		return nil, err
	}

	output, err := u.query.ListProjectTestcases(input.ProjectID)
	if err != nil {
		return nil, err
	}

	return &Output{
		GradingItems: output,
	}, nil
}
