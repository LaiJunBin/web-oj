package getlifecycle_test

import (
	"errors"
	"testing"

	"judge-backend/internal/app/project_domain/project/application/query/dto"
	getlifecycle "judge-backend/internal/app/project_domain/project/application/usecase/lifecycle/get_lifecycle"
	"judge-backend/internal/app/project_domain/project/mocks"

	"github.com/stretchr/testify/assert"
)

type UsecaseMocks struct {
	query *mocks.LifecycleQuery
}

func NewUsecaseMocks(t *testing.T) *UsecaseMocks {
	return &UsecaseMocks{
		query: mocks.NewLifecycleQuery(t),
	}
}

func NewUsecase(t *testing.T) (getlifecycle.Usecase, *UsecaseMocks) {
	mocks := NewUsecaseMocks(t)
	return getlifecycle.NewUsecase(mocks.query), mocks
}

func TestExecute(t *testing.T) {
	t.Run("should get lifecycle successfully", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &getlifecycle.Input{ID: 1}
		lifecycle := &dto.LifecycleCommand{}
		mocks.query.On("GetLifecycle", input.ID).Return(lifecycle, nil).Once()

		output, err := usecase.Execute(input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, lifecycle, output.Lifecycle)
	})

	t.Run("should return error when ID is invalid", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		input := &getlifecycle.Input{ID: 0}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, getlifecycle.ErrorInvalidID, err)
	})

	t.Run("should return error when lifecycle not found", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &getlifecycle.Input{ID: 1}
		mocks.query.On("GetLifecycle", input.ID).Return(nil, nil).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, getlifecycle.ErrorLifecycleNotFound, err)
	})

	t.Run("should return internal server error when query fails", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &getlifecycle.Input{ID: 1}
		mocks.query.On("GetLifecycle", input.ID).Return(nil, errors.New("query error")).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, getlifecycle.ErrorInternalServer, err)
	})
}
