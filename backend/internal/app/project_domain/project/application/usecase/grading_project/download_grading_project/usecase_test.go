package downloadgradingproject_test

import (
	"errors"
	"testing"

	downloadgradingproject "judge-backend/internal/app/project_domain/project/application/usecase/grading_project/download_grading_project"
	gradingproject "judge-backend/internal/app/project_domain/project/application/service/grading_project"
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

func NewUsecase(t *testing.T) (downloadgradingproject.Usecase, *UsecaseMocks) {
	m := NewUsecaseMocks(t)
	return downloadgradingproject.NewUsecase(m.repository, m.service), m
}

func TestExecute(t *testing.T) {
	t.Run("should download grading project successfully", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		input := &downloadgradingproject.Input{ProjectID: 1}
		file := &gradingproject.File{Name: "project.zip", Content: []byte("data"), Size: 4}
		m.repository.On("Exists", int(input.ProjectID)).Return(true, nil).Once()
		m.service.On("DownloadGradingProject", input.ProjectID).Return(file, nil).Once()

		output, err := usecase.Execute(input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, file, output.File)
	})

	t.Run("should return error when project ID is invalid", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		input := &downloadgradingproject.Input{ProjectID: 0}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, downloadgradingproject.ErrorInvalidProjectID, err)
	})

	t.Run("should return ErrorInternalServer when Exists returns error", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		input := &downloadgradingproject.Input{ProjectID: 1}
		m.repository.On("Exists", int(input.ProjectID)).Return(false, errors.New("db error")).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, downloadgradingproject.ErrorInternalServer, err)
	})

	t.Run("should return ErrorProjectNotFound when project does not exist", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		input := &downloadgradingproject.Input{ProjectID: 1}
		m.repository.On("Exists", int(input.ProjectID)).Return(false, nil).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, downloadgradingproject.ErrorProjectNotFound, err)
	})

	// No grading files uploaded is a legitimate state of an existing project:
	// the download answers 404 through the usecase's not-found error, not 500.
	t.Run("should return ErrorProjectNotFound when no grading project was uploaded", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		input := &downloadgradingproject.Input{ProjectID: 1}
		m.repository.On("Exists", int(input.ProjectID)).Return(true, nil).Once()
		m.service.On("DownloadGradingProject", input.ProjectID).Return(nil, gradingproject.ErrorGradingProjectNotFound).Once()

		_, err := usecase.Execute(input)

		assert.ErrorIs(t, err, downloadgradingproject.ErrorProjectNotFound)
	})

	t.Run("should return ErrorInternalServer when download service fails", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		input := &downloadgradingproject.Input{ProjectID: 1}
		m.repository.On("Exists", int(input.ProjectID)).Return(true, nil).Once()
		m.service.On("DownloadGradingProject", input.ProjectID).Return(nil, errors.New("service error")).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, downloadgradingproject.ErrorInternalServer, err)
	})
}
