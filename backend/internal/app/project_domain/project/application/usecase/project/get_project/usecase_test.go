package getproject_test

import (
	"judge-backend/internal/app/project_domain/project/application/query/dto"
	getproject "judge-backend/internal/app/project_domain/project/application/usecase/project/get_project"
	"judge-backend/internal/app/project_domain/project/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
)

type UsecaseMocks struct {
	query *mocks.ProjectQuery
}

func NewUsecaseMocks(t *testing.T) *UsecaseMocks {
	return &UsecaseMocks{
		query: mocks.NewProjectQuery(t),
	}
}

func NewUsecase(t *testing.T) (getproject.Usecase, *UsecaseMocks) {
	mocks := NewUsecaseMocks(t)
	return getproject.NewUsecase(mocks.query), mocks
}

func TestUsecase_Execute(t *testing.T) {
	t.Run("should return project", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		mocks.query.On("GetProjectByID", int32(1)).Return(&dto.ProjectDetail{}, nil).Once()

		output, err := usecase.Execute(&getproject.Input{ID: 1})

		assert.NoError(t, err)
		assert.NotNil(t, output)
	})

	t.Run("should return error when id is invalid", func(t *testing.T) {
		usecase, _ := NewUsecase(t)

		_, err := usecase.Execute(&getproject.Input{ID: 0})

		assert.Error(t, err)
	})

	t.Run("should return internal server error when query returns error", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		mocks.query.On("GetProjectByID", int32(1)).Return(nil, assert.AnError).Once()

		_, err := usecase.Execute(&getproject.Input{ID: 1})

		assert.Error(t, err)
	})
}
