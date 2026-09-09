package updateproject_test

import (
	"judge-backend/internal/app/shared/test"
	updateproject "judge-backend/internal/app/project_domain/project/application/usecase/project/update_project"
	"judge-backend/internal/app/project_domain/project/domain/entity/project"
	"judge-backend/internal/app/project_domain/project/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
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

func NewUsecase(t *testing.T) (updateproject.Usecase, *UsecaseMocks) {
	mocks := NewUsecaseMocks(t)
	return updateproject.NewUsecase(mocks.repository, mocks.imageRepo, mocks.resultParserValidator), mocks
}

func TestExecute(t *testing.T) {
	t.Run("should update project successfully", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &updateproject.Input{
			ID:           1,
			Name:         test.StringPtr("Updated Project"),
			Description:  test.StringPtr("Updated Description"),
			ImageTag:     test.StringPtr("nginx:latest"),
			TestImageTag: test.StringPtr("test-nginx:latest"),
			Env:          test.StringPtr("a=b"),
			TestEnv:      test.StringPtr("c=d"),
			DynamicFieldConfigs: &[]*project.DynamicFieldConfig{{
				Name:       "Updated Field",
				Type:       "text",
				Label:      "Updated Label",
				DefaultVal: test.StringPtr("Updated Default"),
				IsRequired: true,
			}},
		}
		project := &project.Project{}
		mocks.repository.On("FindByID", int(input.ID)).Return(project, nil).Once()
		mocks.imageRepo.On("Exists", *input.ImageTag).Return(true, nil).Once()
		mocks.imageRepo.On("Exists", *input.TestImageTag).Return(true, nil).Once()
		mocks.repository.On("Save", project).Return(nil).Once()

		output, err := usecase.Execute(input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
	})

	t.Run("should return error when no input provided", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		input := &updateproject.Input{ID: 1}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, updateproject.ErrorNoInput, err)
	})

	t.Run("should return error when name is empty", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		input := &updateproject.Input{
			ID:   1,
			Name: test.StringPtr(""),
		}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, updateproject.ErrorNameIsRequired, err)
	})

	t.Run("should return error when image tag is empty", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		input := &updateproject.Input{
			ID:       1,
			ImageTag: test.StringPtr(""),
		}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, updateproject.ErrorImageTagIsRequired, err)
	})

	t.Run("should return error when test image tag is empty", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		input := &updateproject.Input{
			ID:           1,
			TestImageTag: test.StringPtr(""),
		}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, updateproject.ErrorTestImageTagIsRequired, err)
	})

	t.Run("should return error when project not found", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &updateproject.Input{ID: 1, Name: test.StringPtr("Updated Project")}
		mocks.repository.On("FindByID", int(input.ID)).Return(nil, nil).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, updateproject.ErrorProjectNotFound, err)
	})

	t.Run("should return internal server error when repository find fails", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &updateproject.Input{ID: 1, Name: test.StringPtr("Updated Project")}
		mocks.repository.On("FindByID", int(input.ID)).Return(nil, assert.AnError).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, updateproject.ErrorInternalServer, err)
	})

	t.Run("should return error when image not found", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &updateproject.Input{
			ID:       1,
			ImageTag: test.StringPtr("nginx:latest"),
		}
		project := &project.Project{}
		mocks.repository.On("FindByID", int(input.ID)).Return(project, nil).Once()
		mocks.imageRepo.On("Exists", *input.ImageTag).Return(false, nil).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, updateproject.ErrorImageNotFound, err)
	})

	t.Run("should return error when test image not found", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &updateproject.Input{
			ID:           1,
			TestImageTag: test.StringPtr("test-nginx:latest"),
		}
		project := &project.Project{}
		mocks.repository.On("FindByID", int(input.ID)).Return(project, nil).Once()
		mocks.imageRepo.On("Exists", *input.TestImageTag).Return(false, nil).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, updateproject.ErrorTestImageNotFound, err)
	})

	t.Run("should return internal server error when repository save fails", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &updateproject.Input{
			ID:           1,
			Name:         test.StringPtr("Updated Project"),
			Description:  test.StringPtr("Updated Description"),
			ImageTag:     test.StringPtr("nginx:latest"),
			TestImageTag: test.StringPtr("test-nginx:latest"),
			Env:          test.StringPtr("a=b"),
			TestEnv:      test.StringPtr("c=d"),
			DynamicFieldConfigs: &[]*project.DynamicFieldConfig{{
				Name:       "Updated Field",
				Type:       "text",
				Label:      "Updated Label",
				DefaultVal: test.StringPtr("Updated Default"),
				IsRequired: true,
			}},
		}
		project := &project.Project{}
		mocks.repository.On("FindByID", int(input.ID)).Return(project, nil).Once()
		mocks.imageRepo.On("Exists", *input.ImageTag).Return(true, nil).Once()
		mocks.imageRepo.On("Exists", *input.TestImageTag).Return(true, nil).Once()
		mocks.repository.On("Save", project).Return(assert.AnError).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, updateproject.ErrorInternalServer, err)
	})
}
