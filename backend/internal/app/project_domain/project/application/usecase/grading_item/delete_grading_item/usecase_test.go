package deletegradingitem_test

import (
	"errors"
	"testing"

	deletegradingitem "judge-backend/internal/app/project_domain/project/application/usecase/grading_item/delete_grading_item"
	"judge-backend/internal/app/project_domain/project/domain/entity/gradingitem"
	"judge-backend/internal/app/project_domain/project/mocks"

	"github.com/stretchr/testify/assert"
)

type UsecaseMocks struct {
	repository *mocks.GradingItemRepository
}

func NewUsecaseMocks(t *testing.T) *UsecaseMocks {
	return &UsecaseMocks{
		repository: mocks.NewGradingItemRepository(t),
	}
}

func NewUsecase(t *testing.T) (deletegradingitem.Usecase, *UsecaseMocks) {
	mocks := NewUsecaseMocks(t)
	return deletegradingitem.NewUsecase(mocks.repository), mocks
}

func TestExecute(t *testing.T) {
	t.Run("should delete grading item successfully", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &deletegradingitem.Input{ID: 1}
		gradingItem := &gradingitem.GradingItem{}
		mocks.repository.On("FindByID", int(input.ID)).Return(gradingItem, nil).Once()
		mocks.repository.On("Delete", gradingItem).Return(nil).Once()

		output, err := usecase.Execute(input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
	})

	t.Run("should return error when ID is invalid", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		input := &deletegradingitem.Input{ID: 0}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, deletegradingitem.ErrorInvalidID, err)
	})

	t.Run("should return error when grading item not found", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &deletegradingitem.Input{ID: 1}
		mocks.repository.On("FindByID", int(input.ID)).Return(nil, nil).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, deletegradingitem.ErrorGradingItemNotFound, err)
	})

	t.Run("should return internal server error when repository find fails", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &deletegradingitem.Input{ID: 1}
		mocks.repository.On("FindByID", int(input.ID)).Return(nil, errors.New("find error")).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, deletegradingitem.ErrorInternalServer, err)
	})

	t.Run("should return internal server error when repository delete fails", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &deletegradingitem.Input{ID: 1}
		gradingItem := &gradingitem.GradingItem{}
		mocks.repository.On("FindByID", int(input.ID)).Return(gradingItem, nil).Once()
		mocks.repository.On("Delete", gradingItem).Return(errors.New("delete error")).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, deletegradingitem.ErrorInternalServer, err)
	})
}
