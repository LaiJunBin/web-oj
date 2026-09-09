package readgradingprojectfile_test

import (
	"bytes"
	"errors"
	"testing"

	readgradingprojectfile "judge-backend/internal/app/project_domain/project/application/usecase/grading_project/read_grading_project_file"
	"judge-backend/internal/app/project_domain/project/mocks"

	"github.com/stretchr/testify/assert"
)

type UsecaseMocks struct {
	repository *mocks.ProjectRepository
	service    *mocks.GradingProjectService
	base64     *mocks.Base64Service
}

func NewUsecaseMocks(t *testing.T) *UsecaseMocks {
	return &UsecaseMocks{
		repository: mocks.NewProjectRepository(t),
		service:    mocks.NewGradingProjectService(t),
		base64:     mocks.NewBase64Service(t),
	}
}

func NewUsecase(t *testing.T) (readgradingprojectfile.Usecase, *UsecaseMocks) {
	m := NewUsecaseMocks(t)
	return readgradingprojectfile.NewUsecase(m.repository, m.service, m.base64), m
}

func TestExecute(t *testing.T) {
	t.Run("should read grading project file successfully", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		input := &readgradingprojectfile.Input{ProjectID: 1, Base64EncodedPath: "bWFpbi5nbw=="}
		decodedPath := "main.go"
		fileContent := "package main"
		fileBuffer := bytes.NewBufferString(fileContent)
		encodedContent := "cGFja2FnZSBtYWlu"
		m.repository.On("Exists", int(input.ProjectID)).Return(true, nil).Once()
		m.base64.On("Decode", input.Base64EncodedPath).Return(decodedPath, nil).Once()
		m.service.On("ReadFile", input.ProjectID, decodedPath).Return(fileBuffer, nil).Once()
		m.base64.On("Encode", fileContent).Return(encodedContent).Once()

		output, err := usecase.Execute(input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, encodedContent, output.Base64Content)
	})

	t.Run("should return error when project ID is invalid", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		input := &readgradingprojectfile.Input{ProjectID: 0, Base64EncodedPath: "bWFpbi5nbw=="}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, readgradingprojectfile.ErrorInvalidProjectID, err)
	})

	t.Run("should return error when base64 encoded path is empty", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		input := &readgradingprojectfile.Input{ProjectID: 1, Base64EncodedPath: ""}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, readgradingprojectfile.ErrorBase64EncodedPathRequired, err)
	})

	t.Run("should return ErrorInternalServer when Exists returns error", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		input := &readgradingprojectfile.Input{ProjectID: 1, Base64EncodedPath: "bWFpbi5nbw=="}
		m.repository.On("Exists", int(input.ProjectID)).Return(false, errors.New("db error")).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, readgradingprojectfile.ErrorInternalServer, err)
	})

	t.Run("should return ErrorProjectNotFound when project does not exist", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		input := &readgradingprojectfile.Input{ProjectID: 1, Base64EncodedPath: "bWFpbi5nbw=="}
		m.repository.On("Exists", int(input.ProjectID)).Return(false, nil).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, readgradingprojectfile.ErrorProjectNotFound, err)
	})

	t.Run("should return ErrorInvalidBase64EncodedPath when decode fails", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		input := &readgradingprojectfile.Input{ProjectID: 1, Base64EncodedPath: "!!!invalid!!!"}
		m.repository.On("Exists", int(input.ProjectID)).Return(true, nil).Once()
		m.base64.On("Decode", input.Base64EncodedPath).Return("", errors.New("decode error")).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, readgradingprojectfile.ErrorInvalidBase64EncodedPath, err)
	})

	t.Run("should return ErrorInternalServer when ReadFile fails", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		input := &readgradingprojectfile.Input{ProjectID: 1, Base64EncodedPath: "bWFpbi5nbw=="}
		decodedPath := "main.go"
		m.repository.On("Exists", int(input.ProjectID)).Return(true, nil).Once()
		m.base64.On("Decode", input.Base64EncodedPath).Return(decodedPath, nil).Once()
		m.service.On("ReadFile", input.ProjectID, decodedPath).Return(nil, errors.New("read error")).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, readgradingprojectfile.ErrorInternalServer, err)
	})
}
