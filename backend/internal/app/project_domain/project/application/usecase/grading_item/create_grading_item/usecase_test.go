package creategradingitem_test

import (
	"errors"
	"testing"

	creategradingitem "judge-backend/internal/app/project_domain/project/application/usecase/grading_item/create_grading_item"
	"judge-backend/internal/app/project_domain/project/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type UsecaseMocks struct {
	repository            *mocks.GradingItemRepository
	projectRepo           *mocks.ProjectRepository
	resultParserValidator *mocks.ResultParserValidator
}

func NewUsecaseMocks(t *testing.T) *UsecaseMocks {
	return &UsecaseMocks{
		repository:            mocks.NewGradingItemRepository(t),
		projectRepo:           mocks.NewProjectRepository(t),
		resultParserValidator: mocks.NewResultParserValidator(t),
	}
}

func NewUsecase(t *testing.T) (creategradingitem.Usecase, *UsecaseMocks) {
	mocks := NewUsecaseMocks(t)
	return creategradingitem.NewUsecase(mocks.repository, mocks.projectRepo, mocks.resultParserValidator), mocks
}

func TestExecute(t *testing.T) {
	t.Run("should create grading item successfully", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &creategradingitem.Input{
			ProjectID:       1,
			Name:            "Test Grading Item",
			Label:           "1",
			Description:     "Description",
			Score:           10,
			ScoringStrategy: "increase",
			OutputFormat:    "text",
			DependTestcases: []int32{1, 2},
		}
		mocks.projectRepo.On("Exists", int(input.ProjectID)).Return(true, nil).Once()
		mocks.repository.On("Save", mock.AnythingOfType("*gradingitem.GradingItem")).Return(nil).Once()

		output, err := usecase.Execute(input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
	})

	t.Run("should return error when project ID is invalid", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		input := &creategradingitem.Input{ProjectID: 0}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, creategradingitem.ErrorInvalidProjectID, err)
	})

	t.Run("should return error when project not found", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &creategradingitem.Input{
			ProjectID:       1,
			Name:            "Test Grading Item",
			Label:           "1",
			Description:     "Description",
			Score:           10,
			ScoringStrategy: "increase",
			OutputFormat:    "text",
			DependTestcases: []int32{1, 2},
		}
		mocks.projectRepo.On("Exists", int(input.ProjectID)).Return(false, nil).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, creategradingitem.ErrorProjectNotFound, err)
	})

	t.Run("should return internal server error when project repo fails", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &creategradingitem.Input{
			ProjectID:       1,
			Name:            "Test Grading Item",
			Label:           "1",
			Description:     "Description",
			Score:           10,
			ScoringStrategy: "increase",
			OutputFormat:    "text",
			DependTestcases: []int32{1, 2},
		}
		mocks.projectRepo.On("Exists", int(input.ProjectID)).Return(false, errors.New("repo error")).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, creategradingitem.ErrorInternalServer, err)
	})

	t.Run("should return error when grading item save fails", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &creategradingitem.Input{
			ProjectID:       1,
			Name:            "Test Grading Item",
			Label:           "1",
			Description:     "Description",
			Score:           10,
			ScoringStrategy: "increase",
			OutputFormat:    "text",
			DependTestcases: []int32{1, 2},
		}
		mocks.projectRepo.On("Exists", int(input.ProjectID)).Return(true, nil).Once()
		mocks.repository.On("Save", mock.AnythingOfType("*gradingitem.GradingItem")).Return(errors.New("save error")).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, creategradingitem.ErrorInternalServer, err)
	})

	t.Run("should return error when name is empty", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		input := &creategradingitem.Input{
			ProjectID:       1,
			Name:            "",
			Description:     "Description",
			Score:           10,
			ScoringStrategy: "increase",
			OutputFormat:    "text",
			DependTestcases: []int32{1, 2},
		}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, creategradingitem.ErrorEmptyName, err)
	})

	t.Run("should return error when label is empty", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		input := &creategradingitem.Input{
			ProjectID:       1,
			Name:            "Test Grading Item",
			Label:           "",
			Description:     "Description",
			Score:           10,
			ScoringStrategy: "increase",
			OutputFormat:    "text",
			DependTestcases: []int32{1, 2},
		}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, creategradingitem.ErrorEmptyLabel, err)
	})

	t.Run("should return error when score is invalid", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		input := &creategradingitem.Input{
			ProjectID:       1,
			Name:            "Test Grading Item",
			Label:           "1",
			Description:     "Description",
			Score:           -1,
			ScoringStrategy: "increase",
			OutputFormat:    "text",
			DependTestcases: []int32{1, 2},
		}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, creategradingitem.ErrorInvalidScore, err)
	})

	t.Run("should return error when scoring strategy is empty", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		input := &creategradingitem.Input{
			ProjectID:       1,
			Name:            "Test Grading Item",
			Label:           "1",
			Description:     "Description",
			Score:           10,
			ScoringStrategy: "",
			OutputFormat:    "text",
			DependTestcases: []int32{1, 2},
		}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, creategradingitem.ErrorEmptyScoringStrategy, err)
	})

	t.Run("should return error when output formatter is empty", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		input := &creategradingitem.Input{
			ProjectID:       1,
			Name:            "Test Grading Item",
			Label:           "1",
			Description:     "Description",
			Score:           10,
			ScoringStrategy: "increase",
			OutputFormat:    "",
			DependTestcases: []int32{1, 2},
		}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, creategradingitem.ErrorEmptyOutputFormat, err)
	})
}
