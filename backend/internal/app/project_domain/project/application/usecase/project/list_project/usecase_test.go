package listproject_test

import (
	"judge-backend/internal/app/project_domain/project/application/query/dto"
	listproject "judge-backend/internal/app/project_domain/project/application/usecase/project/list_project"
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

func NewUsecase(t *testing.T) (listproject.Usecase, *UsecaseMocks) {
	mocks := NewUsecaseMocks(t)
	return listproject.NewUsecase(mocks.query), mocks
}

func TestExecute(t *testing.T) {
	t.Run("should return list of projects", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		mocks.query.On("ListProject").Return([]dto.Project{}, nil).Once()

		output, err := usecase.Execute(&listproject.Input{})

		assert.NoError(t, err)
		assert.NotNil(t, output)
	})

	t.Run("should return internal server error when query returns error", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		mocks.query.On("ListProject").Return(nil, assert.AnError).Once()

		_, err := usecase.Execute(&listproject.Input{})

		assert.Error(t, err)
	})
}
