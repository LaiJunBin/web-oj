package deleteproject_test

import (
	deleteproject "judge-backend/internal/app/project_domain/project/application/usecase/project/delete_project"
	"judge-backend/internal/app/project_domain/project/domain/entity/project"
	"judge-backend/internal/app/project_domain/project/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
)

type UsecaseMocks struct {
	repository *mocks.ProjectRepository
}

func NewUsecaseMocks(t *testing.T) *UsecaseMocks {
	return &UsecaseMocks{
		repository: mocks.NewProjectRepository(t),
	}
}

func NewUsecase(t *testing.T) (deleteproject.Usecase, *UsecaseMocks) {
	mocks := NewUsecaseMocks(t)
	return deleteproject.NewUsecase(mocks.repository), mocks
}

func TestExecute(t *testing.T) {
	t.Run("should delete project successfully", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &deleteproject.Input{ID: 1}
		project := &project.Project{}
		mocks.repository.On("FindByID", int(input.ID)).Return(project, nil).Once()
		mocks.repository.On("Delete", project).Return(nil).Once()

		output, err := usecase.Execute(input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
	})

	t.Run("should return error when id is invalid", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		input := &deleteproject.Input{ID: 0}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, deleteproject.ErrorInvalidID, err)
	})

	t.Run("should return error when project not found", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &deleteproject.Input{ID: 1}
		mocks.repository.On("FindByID", int(input.ID)).Return(nil, nil).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, deleteproject.ErrorProjectNotFound, err)
	})

	t.Run("should return internal server error when repository find fails", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &deleteproject.Input{ID: 1}
		mocks.repository.On("FindByID", int(input.ID)).Return(nil, assert.AnError).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, deleteproject.ErrorInternalServer, err)
	})

	t.Run("should return internal server error when repository delete fails", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &deleteproject.Input{ID: 1}
		project := &project.Project{}
		mocks.repository.On("FindByID", int(input.ID)).Return(project, nil).Once()
		mocks.repository.On("Delete", project).Return(assert.AnError).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, deleteproject.ErrorInternalServer, err)
	})
}
