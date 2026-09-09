package movegradingitemorder_test

import (
	"errors"
	"testing"

	movegradingitemorder "judge-backend/internal/app/project_domain/project/application/usecase/grading_item/move_grading_item_order"
	"judge-backend/internal/app/project_domain/project/domain/entity/gradingitem"
	"judge-backend/internal/app/project_domain/project/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type UsecaseMocks struct {
	repository *mocks.GradingItemRepository
}

func NewUsecaseMocks(t *testing.T) *UsecaseMocks {
	return &UsecaseMocks{
		repository: mocks.NewGradingItemRepository(t),
	}
}

func NewUsecase(t *testing.T) (movegradingitemorder.Usecase, *UsecaseMocks) {
	m := NewUsecaseMocks(t)
	return movegradingitemorder.NewUsecase(m.repository), m
}

func TestExecute(t *testing.T) {
	t.Run("should return error when ID is invalid", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		input := &movegradingitemorder.Input{ID: 0, Order: 1}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, movegradingitemorder.ErrorInvalidID, err)
	})

	t.Run("should return error when order is invalid", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		input := &movegradingitemorder.Input{ID: 1, Order: 0}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, movegradingitemorder.ErrorInvalidOrder, err)
	})

	t.Run("should return ErrorInternalServer when FindByID returns error", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		input := &movegradingitemorder.Input{ID: 1, Order: 2}
		m.repository.On("FindByID", int(input.ID)).Return(nil, errors.New("db error")).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, movegradingitemorder.ErrorInternalServer, err)
	})

	t.Run("should return ErrorGradingItemNotFound when FindByID returns nil", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		input := &movegradingitemorder.Input{ID: 1, Order: 2}
		m.repository.On("FindByID", int(input.ID)).Return(nil, nil).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, movegradingitemorder.ErrorGradingItemNotFound, err)
	})

	t.Run("should return ErrorInternalServer when Save returns error", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		input := &movegradingitemorder.Input{ID: 1, Order: 2}
		item := &gradingitem.GradingItem{}
		m.repository.On("FindByID", int(input.ID)).Return(item, nil).Once()
		m.repository.On("Save", mock.AnythingOfType("*gradingitem.GradingItem")).Return(errors.New("save error")).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, movegradingitemorder.ErrorInternalServer, err)
	})

	t.Run("should move grading item order successfully", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		input := &movegradingitemorder.Input{ID: 1, Order: 3}
		item := &gradingitem.GradingItem{}
		m.repository.On("FindByID", int(input.ID)).Return(item, nil).Once()
		m.repository.On("Save", mock.AnythingOfType("*gradingitem.GradingItem")).Return(nil).Once()

		output, err := usecase.Execute(input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
	})
}
