package deletelifecycle_test

import (
	"errors"
	"testing"

	deletelifecycle "judge-backend/internal/app/project_domain/project/application/usecase/lifecycle/delete_lifecycle"
	"judge-backend/internal/app/project_domain/project/domain/entity/lifecycle"
	"judge-backend/internal/app/project_domain/project/mocks"

	"github.com/stretchr/testify/assert"
)

type UsecaseMocks struct {
	repository *mocks.LifecycleRepository
}

func NewUsecaseMocks(t *testing.T) *UsecaseMocks {
	return &UsecaseMocks{
		repository: mocks.NewLifecycleRepository(t),
	}
}

func NewUsecase(t *testing.T) (deletelifecycle.Usecase, *UsecaseMocks) {
	mocks := NewUsecaseMocks(t)
	return deletelifecycle.NewUsecase(mocks.repository), mocks
}

func TestExecute(t *testing.T) {
	t.Run("should delete lifecycle successfully", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &deletelifecycle.Input{ID: 1}
		lifecycle := &lifecycle.Lifecycle{}
		mocks.repository.On("FindByID", int(input.ID)).Return(lifecycle, nil).Once()
		mocks.repository.On("Delete", lifecycle).Return(nil).Once()

		output, err := usecase.Execute(input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
	})

	t.Run("should return error when ID is invalid", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		input := &deletelifecycle.Input{ID: 0}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, deletelifecycle.ErrorInvalidID, err)
	})

	t.Run("should return error when lifecycle not found", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &deletelifecycle.Input{ID: 1}
		mocks.repository.On("FindByID", int(input.ID)).Return(nil, nil).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, deletelifecycle.ErrorLifecycleNotFound, err)
	})

	t.Run("should return internal server error when repository find fails", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &deletelifecycle.Input{ID: 1}
		mocks.repository.On("FindByID", int(input.ID)).Return(nil, errors.New("find error")).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, deletelifecycle.ErrorInternalServer, err)
	})

	t.Run("should return internal server error when repository delete fails", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &deletelifecycle.Input{ID: 1}
		lifecycle := &lifecycle.Lifecycle{}
		mocks.repository.On("FindByID", int(input.ID)).Return(lifecycle, nil).Once()
		mocks.repository.On("Delete", lifecycle).Return(errors.New("delete error")).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, deletelifecycle.ErrorInternalServer, err)
	})
}
