package updatelifecycle_test

import (
	"errors"
	"testing"

	"judge-backend/internal/app/shared/test"
	updatelifecycle "judge-backend/internal/app/project_domain/project/application/usecase/lifecycle/update_lifecycle"
	"judge-backend/internal/app/project_domain/project/domain/entity/lifecycle"
	"judge-backend/internal/app/project_domain/project/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type UsecaseMocks struct {
	repository *mocks.LifecycleRepository
}

func NewUsecaseMocks(t *testing.T) *UsecaseMocks {
	return &UsecaseMocks{
		repository: mocks.NewLifecycleRepository(t),
	}
}

func NewUsecase(t *testing.T) (updatelifecycle.Usecase, *UsecaseMocks) {
	mocks := NewUsecaseMocks(t)
	return updatelifecycle.NewUsecase(mocks.repository), mocks
}

func TestExecute(t *testing.T) {
	t.Run("should update lifecycle successfully", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &updatelifecycle.Input{
			ID:                   1,
			Type:                 test.StringPtr("before_all"),
			Name:                 test.StringPtr("Updated Lifecycle"),
			Description:          test.StringPtr("Updated Description"),
			Command:              test.StringPtr("Updated Command"),
			ExecutionEnvironment: test.StringPtr("project"),
		}
		lifecycle := &lifecycle.Lifecycle{}
		mocks.repository.On("FindByID", int(input.ID)).Return(lifecycle, nil).Once()
		mocks.repository.On("Save", mock.AnythingOfType("*lifecycle.Lifecycle")).Return(nil).Once()

		output, err := usecase.Execute(input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
	})

	t.Run("should return error when ID is invalid", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		input := &updatelifecycle.Input{ID: 0}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, updatelifecycle.ErrorInvalidID, err)
	})

	t.Run("should return error when lifecycle not found", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &updatelifecycle.Input{
			ID:                   1,
			Type:                 test.StringPtr("before_all"),
			Name:                 test.StringPtr("Updated Lifecycle"),
			Description:          test.StringPtr("Updated Description"),
			Command:              test.StringPtr("Updated Command"),
			ExecutionEnvironment: test.StringPtr("project"),
		}
		mocks.repository.On("FindByID", int(input.ID)).Return(nil, nil).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, updatelifecycle.ErrorLifecycleNotFound, err)
	})

	t.Run("should return internal server error when repository find fails", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &updatelifecycle.Input{
			ID:                   1,
			Type:                 test.StringPtr("before_all"),
			Name:                 test.StringPtr("Updated Lifecycle"),
			Description:          test.StringPtr("Updated Description"),
			Command:              test.StringPtr("Updated Command"),
			ExecutionEnvironment: test.StringPtr("project"),
		}
		mocks.repository.On("FindByID", int(input.ID)).Return(nil, errors.New("find error")).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, updatelifecycle.ErrorInternalServer, err)
	})

	t.Run("should return error when name is empty", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &updatelifecycle.Input{
			ID:                   1,
			Type:                 test.StringPtr("before_all"),
			Name:                 test.StringPtr(""),
			Description:          test.StringPtr("Updated Description"),
			Command:              test.StringPtr("Updated Command"),
			ExecutionEnvironment: test.StringPtr("project"),
		}
		mocks.repository.On("FindByID", int(input.ID)).Return(&lifecycle.Lifecycle{}, nil).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, updatelifecycle.ErrorInvalidName, err)
	})

	t.Run("should return error when command is empty", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &updatelifecycle.Input{
			ID:                   1,
			Type:                 test.StringPtr("before_all"),
			Name:                 test.StringPtr("Updated Lifecycle"),
			Description:          test.StringPtr("Updated Description"),
			Command:              test.StringPtr(""),
			ExecutionEnvironment: test.StringPtr("project"),
		}
		mocks.repository.On("FindByID", int(input.ID)).Return(&lifecycle.Lifecycle{}, nil).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, updatelifecycle.ErrorInvalidCommand, err)
	})

	t.Run("should return error when execution environment is empty", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &updatelifecycle.Input{
			ID:                   1,
			Type:                 test.StringPtr("before_all"),
			Name:                 test.StringPtr("Updated Lifecycle"),
			Description:          test.StringPtr("Updated Description"),
			Command:              test.StringPtr("Updated Command"),
			ExecutionEnvironment: test.StringPtr(""),
		}
		mocks.repository.On("FindByID", int(input.ID)).Return(&lifecycle.Lifecycle{}, nil).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, updatelifecycle.ErrorInvalidExecutionEnvironment, err)
	})

	t.Run("should return internal server error when lifecycle save fails", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &updatelifecycle.Input{
			ID:                   1,
			Type:                 test.StringPtr("before_all"),
			Name:                 test.StringPtr("Updated Lifecycle"),
			Description:          test.StringPtr("Updated Description"),
			Command:              test.StringPtr("Updated Command"),
			ExecutionEnvironment: test.StringPtr("project"),
		}
		mocks.repository.On("FindByID", int(input.ID)).Return(&lifecycle.Lifecycle{}, nil).Once()
		mocks.repository.On("Save", mock.AnythingOfType("*lifecycle.Lifecycle")).Return(errors.New("save error")).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, updatelifecycle.ErrorInternalServer, err)
	})
}
