package getgradingitem_test

import (
	"errors"
	"testing"

	"judge-backend/internal/app/project_domain/project/application/query/dto"
	getgradingitem "judge-backend/internal/app/project_domain/project/application/usecase/grading_item/get_grading_item"
	"judge-backend/internal/app/project_domain/project/mocks"

	"github.com/stretchr/testify/assert"
)

type UsecaseMocks struct {
	query *mocks.GradingItemQuery
}

func NewUsecaseMocks(t *testing.T) *UsecaseMocks {
	return &UsecaseMocks{
		query: mocks.NewGradingItemQuery(t),
	}
}

func NewUsecase(t *testing.T) (getgradingitem.Usecase, *UsecaseMocks) {
	mocks := NewUsecaseMocks(t)
	return getgradingitem.NewUsecase(mocks.query), mocks
}

func TestExecute(t *testing.T) {
	t.Run("should get grading item successfully", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &getgradingitem.Input{ID: 1}
		gradingItem := &dto.GradingItemDetail{}
		mocks.query.On("GetGradingItem", input.ID).Return(gradingItem, nil).Once()

		output, err := usecase.Execute(input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, gradingItem, output.GradingItem)
	})

	t.Run("should return error when ID is invalid", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		input := &getgradingitem.Input{ID: 0}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, getgradingitem.ErrorInvalidID, err)
	})

	t.Run("should return error when grading item not found", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &getgradingitem.Input{ID: 1}
		mocks.query.On("GetGradingItem", input.ID).Return(nil, nil).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, getgradingitem.ErrorGradingItemNotFound, err)
	})

	t.Run("should return internal server error when query fails", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &getgradingitem.Input{ID: 1}
		mocks.query.On("GetGradingItem", input.ID).Return(nil, errors.New("query error")).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, getgradingitem.ErrorInternalServer, err)
	})
}
