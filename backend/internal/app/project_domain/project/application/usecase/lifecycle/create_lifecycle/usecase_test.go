package createlifecycle_test

import (
	"errors"
	"testing"

	createlifecycle "judge-backend/internal/app/project_domain/project/application/usecase/lifecycle/create_lifecycle"
	"judge-backend/internal/app/project_domain/project/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type UsecaseMocks struct {
	repository  *mocks.LifecycleRepository
	projectRepo *mocks.ProjectRepository
}

func NewUsecaseMocks(t *testing.T) *UsecaseMocks {
	return &UsecaseMocks{
		repository:  mocks.NewLifecycleRepository(t),
		projectRepo: mocks.NewProjectRepository(t),
	}
}

func NewUsecase(t *testing.T) (createlifecycle.Usecase, *UsecaseMocks) {
	mocks := NewUsecaseMocks(t)
	return createlifecycle.NewUsecase(mocks.repository, mocks.projectRepo), mocks
}

func TestExecute(t *testing.T) {
	t.Run("should create lifecycle successfully", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &createlifecycle.Input{
			ProjectID:            1,
			Type:                 "before_all",
			Name:                 "Test Lifecycle",
			Description:          "Description",
			Command:              "Command",
			ExecutionEnvironment: "project",
		}
		mocks.projectRepo.On("Exists", int(input.ProjectID)).Return(true, nil).Once()
		mocks.repository.On("Save", mock.AnythingOfType("*lifecycle.Lifecycle")).Return(nil).Once()

		output, err := usecase.Execute(input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
	})

	t.Run("should return error when project ID is invalid", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		input := &createlifecycle.Input{ProjectID: 0}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, createlifecycle.ErrorInvalidProjectID, err)
	})

	t.Run("should return error when project not found", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &createlifecycle.Input{
			ProjectID:            1,
			Type:                 "before_all",
			Name:                 "Test Lifecycle",
			Description:          "Description",
			Command:              "Command",
			ExecutionEnvironment: "project",
		}
		mocks.projectRepo.On("Exists", int(input.ProjectID)).Return(false, nil).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, createlifecycle.ErrorProjectNotFound, err)
	})

	t.Run("should return internal server error when project repo fails", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &createlifecycle.Input{
			ProjectID:            1,
			Type:                 "before_all",
			Name:                 "Test Lifecycle",
			Description:          "Description",
			Command:              "Command",
			ExecutionEnvironment: "project",
		}
		mocks.projectRepo.On("Exists", int(input.ProjectID)).Return(false, errors.New("repo error")).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, createlifecycle.ErrorInternalServer, err)
	})

	t.Run("should return error when name is empty", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		input := &createlifecycle.Input{
			ProjectID:            1,
			Type:                 "before_all",
			Name:                 "",
			Description:          "Description",
			Command:              "Command",
			ExecutionEnvironment: "project",
		}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, createlifecycle.ErrorEmptyName, err)
	})

	t.Run("should return error when command is empty", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		input := &createlifecycle.Input{
			ProjectID:            1,
			Type:                 "before_all",
			Name:                 "Test Lifecycle",
			Description:          "Description",
			Command:              "",
			ExecutionEnvironment: "project",
		}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, createlifecycle.ErrorEmptyCommand, err)
	})

	t.Run("should return error when execution environment is empty", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		input := &createlifecycle.Input{
			ProjectID:            1,
			Type:                 "before_all",
			Name:                 "Test Lifecycle",
			Description:          "Description",
			Command:              "Command",
			ExecutionEnvironment: "",
		}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, createlifecycle.ErrorEmptyExecutionEnvironment, err)
	})

	t.Run("should return error when lifecycle save fails", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &createlifecycle.Input{
			ProjectID:            1,
			Type:                 "before_all",
			Name:                 "Test Lifecycle",
			Description:          "Description",
			Command:              "Command",
			ExecutionEnvironment: "project",
		}
		mocks.projectRepo.On("Exists", int(input.ProjectID)).Return(true, nil).Once()
		mocks.repository.On("Save", mock.AnythingOfType("*lifecycle.Lifecycle")).Return(errors.New("save error")).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, createlifecycle.ErrorInternalServer, err)
	})
}
