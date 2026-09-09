package putgradingproject_test

import (
	"bytes"
	"errors"
	"testing"

	putgradingproject "judge-backend/internal/app/project_domain/project/application/usecase/grading_project/put_grading_project"
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

func NewUsecase(t *testing.T) (putgradingproject.Usecase, *UsecaseMocks) {
	m := NewUsecaseMocks(t)
	return putgradingproject.NewUsecase(m.repository, m.service), m
}

func TestExecute(t *testing.T) {
	t.Run("should upload grading project successfully", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		zipProject := bytes.NewBufferString("zip content")
		input := &putgradingproject.Input{ProjectID: 1, ZipProject: zipProject}
		m.repository.On("Exists", int(input.ProjectID)).Return(true, nil).Once()
		m.service.On("UploadGradingProject", input.ProjectID, zipProject).Return(nil).Once()

		output, err := usecase.Execute(input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
	})

	t.Run("should return error when project ID is invalid", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		input := &putgradingproject.Input{ProjectID: 0, ZipProject: bytes.NewBufferString("data")}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, putgradingproject.ErrorInvalidProjectID, err)
	})

	t.Run("should return error when zip project is nil", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		input := &putgradingproject.Input{ProjectID: 1, ZipProject: nil}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, putgradingproject.ErrorZipProjectRequired, err)
	})

	t.Run("should return ErrorInternalServer when Exists returns error", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		input := &putgradingproject.Input{ProjectID: 1, ZipProject: bytes.NewBufferString("data")}
		m.repository.On("Exists", int(input.ProjectID)).Return(false, errors.New("db error")).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, putgradingproject.ErrorInternalServer, err)
	})

	t.Run("should return ErrorProjectNotFound when project does not exist", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		input := &putgradingproject.Input{ProjectID: 1, ZipProject: bytes.NewBufferString("data")}
		m.repository.On("Exists", int(input.ProjectID)).Return(false, nil).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, putgradingproject.ErrorProjectNotFound, err)
	})

	t.Run("should return ErrorInternalServer when upload service fails", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		zipProject := bytes.NewBufferString("zip content")
		input := &putgradingproject.Input{ProjectID: 1, ZipProject: zipProject}
		m.repository.On("Exists", int(input.ProjectID)).Return(true, nil).Once()
		m.service.On("UploadGradingProject", input.ProjectID, zipProject).Return(errors.New("upload error")).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, putgradingproject.ErrorInternalServer, err)
	})
}
