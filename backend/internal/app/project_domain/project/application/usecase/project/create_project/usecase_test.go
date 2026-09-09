package createproject_test

import (
	"judge-backend/internal/app/shared/test"
	createproject "judge-backend/internal/app/project_domain/project/application/usecase/project/create_project"
	"judge-backend/internal/app/project_domain/project/domain/entity/project"
	"judge-backend/internal/app/project_domain/project/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type UsecaseMocks struct {
	repository            *mocks.ProjectRepository
	imageRepo             *mocks.ImageRepository
	resultParserValidator *mocks.ResultParserValidator
}

func NewUsecaseMocks(t *testing.T) *UsecaseMocks {
	return &UsecaseMocks{
		repository:            mocks.NewProjectRepository(t),
		imageRepo:             mocks.NewImageRepository(t),
		resultParserValidator: mocks.NewResultParserValidator(t),
	}
}

func NewUsecase(t *testing.T) (createproject.Usecase, *UsecaseMocks) {
	mocks := NewUsecaseMocks(t)
	return createproject.NewUsecase(mocks.repository, mocks.imageRepo, mocks.resultParserValidator), mocks
}

func TestExecute(t *testing.T) {
	t.Run("should create project successfully", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &createproject.Input{
			Name:         "Test Project",
			Description:  "Test Description",
			ImageTag:     "nginx:latest",
			TestImageTag: "test-nginx:latest",
			Env:          "a=b",
			TestEnv:      "c=d",
			DynamicFieldConfigs: []*project.DynamicFieldConfig{{
				Name:       "Test Field",
				Type:       "text",
				Label:      "Test Label",
				DefaultVal: test.StringPtr("Test Default"),
				IsRequired: true,
			}},
		}
		mocks.imageRepo.On("Exists", input.ImageTag).Return(true, nil).Once()
		mocks.imageRepo.On("Exists", input.TestImageTag).Return(true, nil).Once()
		mocks.repository.On("Save", mock.Anything).Return(nil).Once()

		output, err := usecase.Execute(input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
	})

	t.Run("should return error when name is empty", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		input := &createproject.Input{
			Name:         "",
			ImageTag:     "v1.0",
			TestImageTag: "v1.0-test",
		}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, createproject.ErrorNameIsRequired, err)
	})

	t.Run("should return error when image tag is empty", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		input := &createproject.Input{
			Name:         "Test Project",
			ImageTag:     "",
			TestImageTag: "v1.0-test",
		}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, createproject.ErrorImageTagIsRequired, err)
	})

	t.Run("should return error when test image tag is empty", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		input := &createproject.Input{
			Name:         "Test Project",
			ImageTag:     "v1.0",
			TestImageTag: "",
		}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, createproject.ErrorTestImageTagIsRequired, err)
	})

	t.Run("should return error when image not found", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &createproject.Input{
			Name:                "Test Project",
			Description:         "Test Description",
			ImageTag:            "nginx:latest",
			TestImageTag:        "test-nginx:latest",
			Env:                 "a=b",
			TestEnv:             "c=d",
			DynamicFieldConfigs: []*project.DynamicFieldConfig{},
		}
		mocks.imageRepo.On("Exists", input.ImageTag).Return(false, nil).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, createproject.ErrorImageNotFound, err)
	})

	t.Run("should return error when test image not found", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &createproject.Input{
			Name:                "Test Project",
			Description:         "Test Description",
			ImageTag:            "nginx:latest",
			TestImageTag:        "test-nginx:latest",
			Env:                 "a=b",
			TestEnv:             "c=d",
			DynamicFieldConfigs: []*project.DynamicFieldConfig{},
		}
		mocks.imageRepo.On("Exists", input.ImageTag).Return(true, nil).Once()
		mocks.imageRepo.On("Exists", input.TestImageTag).Return(false, nil).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, createproject.ErrorTestImageNotFound, err)
	})

	t.Run("should return internal server error when repository save fails", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &createproject.Input{
			Name:                "Test Project",
			Description:         "Test Description",
			ImageTag:            "nginx:latest",
			TestImageTag:        "test-nginx:latest",
			Env:                 "a=b",
			TestEnv:             "c=d",
			DynamicFieldConfigs: []*project.DynamicFieldConfig{},
		}
		mocks.imageRepo.On("Exists", input.ImageTag).Return(true, nil).Once()
		mocks.imageRepo.On("Exists", input.TestImageTag).Return(true, nil).Once()
		mocks.repository.On("Save", mock.Anything).Return(assert.AnError).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, createproject.ErrorInternalServer, err)
	})
}
