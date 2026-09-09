package movegradingitemorder

import (
	"errors"
	"judge-backend/internal/app/project_domain/project/domain/repository"
)

var (
	ErrorInvalidID           = errors.New("invalid id, must be greater than 0")
	ErrorInvalidOrder        = errors.New("invalid order, must be greater than 0")
	ErrorGradingItemNotFound = errors.New("grading item not found")
	ErrorInternalServer      = errors.New("internal server error")
)

type Input struct {
	ID    int32
	Order int32
}

type Output struct {
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	repository repository.GradingItemRepository
}

func NewUsecase(
	repository repository.GradingItemRepository,
) Usecase {
	return &usecase{
		repository: repository,
	}
}

func (u *usecase) validate(input *Input) error {
	if input.ID <= 0 {
		return ErrorInvalidID
	}

	if input.Order <= 0 {
		return ErrorInvalidOrder
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

	gradingItem.SetOrder(uint32(input.Order))
	if err := u.repository.Save(gradingItem); err != nil {
		return nil, ErrorInternalServer
	}

	return &Output{}, nil
}
