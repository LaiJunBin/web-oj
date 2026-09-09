package gettestcase

import (
	"errors"
	"judge-backend/internal/app/project_domain/project/application/query"
	"judge-backend/internal/app/project_domain/project/application/query/dto"
)

var (
	ErrorInvalidID        = errors.New("invalid id, must be greater than 0")
	ErrorTestcaseNotFound = errors.New("testcase not found")
	ErrorInternalServer   = errors.New("internal server error")
)

type Input struct {
	ID int32
}

type Output struct {
	Testcase *dto.TestcaseDetail
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
	if input.ID <= 0 {
		return ErrorInvalidID
	}

	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	if err := u.validate(input); err != nil {
		return nil, err
	}

	tc, err := u.query.GetTestcase(input.ID)
	if err != nil {
		return nil, ErrorInternalServer
	}

	if tc == nil {
		return nil, ErrorTestcaseNotFound
	}

	return &Output{
		Testcase: tc,
	}, nil
}
