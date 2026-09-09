package getgradingitem

import (
	"errors"
	"judge-backend/internal/app/project_domain/project/application/query"
	"judge-backend/internal/app/project_domain/project/application/query/dto"
)

var (
	ErrorInvalidID           = errors.New("invalid id, must be greater than 0")
	ErrorGradingItemNotFound = errors.New("grading item not found")
	ErrorInternalServer      = errors.New("internal server error")
)

type Input struct {
	ID int32
}

type Output struct {
	GradingItem *dto.GradingItemDetail
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	query query.GradingItemQuery
}

func NewUsecase(
	query query.GradingItemQuery,
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

	gradingItem, err := u.query.GetGradingItem(input.ID)
	if err != nil {
		return nil, ErrorInternalServer
	}

	if gradingItem == nil {
		return nil, ErrorGradingItemNotFound
	}

	return &Output{GradingItem: gradingItem}, nil
}
