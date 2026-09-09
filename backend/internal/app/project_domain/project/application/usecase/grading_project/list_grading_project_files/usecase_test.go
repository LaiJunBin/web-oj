package listgradingprojectfiles_test

import (
	"errors"
	"testing"

	listgradingprojectfiles "judge-backend/internal/app/project_domain/project/application/usecase/grading_project/list_grading_project_files"
	"judge-backend/internal/app/project_domain/project/mocks"

	"github.com/stretchr/testify/assert"
)

type UsecaseMocks struct {
	repository *mocks.ProjectRepository
	service    *mocks.GradingProjectService
}

func NewUsecaseMocks(t *testing.T) *UsecaseMocks {
	return &UsecaseMocks{
		repository: mocks.NewProjectRepository(t),
		service:    mocks.NewGradingProjectService(t),
	}
}

func NewUsecase(t *testing.T) (listgradingprojectfiles.Usecase, *UsecaseMocks) {
	m := NewUsecaseMocks(t)
	return listgradingprojectfiles.NewUsecase(m.repository, m.service), m
}

func TestExecute(t *testing.T) {
	t.Run("should list grading project files successfully", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		input := &listgradingprojectfiles.Input{ProjectID: 1}
		files := []string{"file1.go", "file2.go"}
		m.repository.On("Exists", int(input.ProjectID)).Return(true, nil).Once()
		m.service.On("ListFile", input.ProjectID).Return(files, nil).Once()

		output, err := usecase.Execute(input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, files, output.Files)
	})

	t.Run("should return error when project ID is invalid", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		input := &listgradingprojectfiles.Input{ProjectID: 0}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, listgradingprojectfiles.ErrorInvalidProjectID, err)
	})

	t.Run("should return ErrorInternalServer when Exists returns error", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		input := &listgradingprojectfiles.Input{ProjectID: 1}
		m.repository.On("Exists", int(input.ProjectID)).Return(false, errors.New("db error")).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, listgradingprojectfiles.ErrorInternalServer, err)
	})

	t.Run("should return ErrorProjectNotFound when project does not exist", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		input := &listgradingprojectfiles.Input{ProjectID: 1}
		m.repository.On("Exists", int(input.ProjectID)).Return(false, nil).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, listgradingprojectfiles.ErrorProjectNotFound, err)
	})

	t.Run("should return ErrorInternalServer when list service fails", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		input := &listgradingprojectfiles.Input{ProjectID: 1}
		m.repository.On("Exists", int(input.ProjectID)).Return(true, nil).Once()
		m.service.On("ListFile", input.ProjectID).Return(nil, errors.New("service error")).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, listgradingprojectfiles.ErrorInternalServer, err)
	})
}
