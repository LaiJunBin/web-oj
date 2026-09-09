package updategradingitem_test

import (
	"errors"
	"testing"

	"judge-backend/internal/app/shared/test"
	updategradingitem "judge-backend/internal/app/project_domain/project/application/usecase/grading_item/update_grading_item"
	"judge-backend/internal/app/project_domain/project/domain/entity/gradingitem"
	"judge-backend/internal/app/project_domain/project/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type UsecaseMocks struct {
	repository            *mocks.GradingItemRepository
	resultParserValidator *mocks.ResultParserValidator
}

func NewUsecaseMocks(t *testing.T) *UsecaseMocks {
	return &UsecaseMocks{
		repository:            mocks.NewGradingItemRepository(t),
		resultParserValidator: mocks.NewResultParserValidator(t),
	}
}

func NewUsecase(t *testing.T) (updategradingitem.Usecase, *UsecaseMocks) {
	mocks := NewUsecaseMocks(t)
	return updategradingitem.NewUsecase(mocks.repository, mocks.resultParserValidator), mocks
}

func TestExecute(t *testing.T) {
	t.Run("should update grading item successfully", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &updategradingitem.Input{
			ID:              1,
			Name:            test.StringPtr("Updated Grading Item"),
			Description:     test.StringPtr("Updated Description"),
			Score:           test.Float64Ptr(20),
			ScoringStrategy: test.StringPtr("decrease"),
			OutputFormat:    test.StringPtr("text"),
			DependTestcases: &[]int32{3, 4},
		}
		gradingItem := &gradingitem.GradingItem{}
		mocks.repository.On("FindByID", int(input.ID)).Return(gradingItem, nil).Once()
		mocks.repository.On("Save", mock.AnythingOfType("*gradingitem.GradingItem")).Return(nil).Once()

		output, err := usecase.Execute(input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
	})

	t.Run("should return error when ID is invalid", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		input := &updategradingitem.Input{ID: 0}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, updategradingitem.ErrorInvalidID, err)
	})

	t.Run("should return error when grading item not found", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &updategradingitem.Input{
			ID:              1,
			Name:            test.StringPtr("Updated Grading Item"),
			Description:     test.StringPtr("Updated Description"),
			Score:           test.Float64Ptr(20),
			ScoringStrategy: test.StringPtr("decrease"),
			OutputFormat:    test.StringPtr("text"),
			DependTestcases: &[]int32{3, 4},
		}
		mocks.repository.On("FindByID", int(input.ID)).Return(nil, nil).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, updategradingitem.ErrorGradingItemNotFound, err)
	})

	t.Run("should return internal server error when repository find fails", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &updategradingitem.Input{
			ID:              1,
			Name:            test.StringPtr("Updated Grading Item"),
			Description:     test.StringPtr("Updated Description"),
			Score:           test.Float64Ptr(20),
			ScoringStrategy: test.StringPtr("decrease"),
			OutputFormat:    test.StringPtr("text"),
			DependTestcases: &[]int32{3, 4},
		}
		mocks.repository.On("FindByID", int(input.ID)).Return(nil, errors.New("find error")).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, updategradingitem.ErrorInternalServer, err)
	})

	t.Run("should return error when name is empty", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &updategradingitem.Input{
			ID:              1,
			Name:            test.StringPtr(""),
			Description:     test.StringPtr("Updated Description"),
			Score:           test.Float64Ptr(20),
			ScoringStrategy: test.StringPtr("decrease"),
			OutputFormat:    test.StringPtr("text"),
			DependTestcases: &[]int32{3, 4},
		}
		mocks.repository.On("FindByID", int(input.ID)).Return(&gradingitem.GradingItem{}, nil).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, updategradingitem.ErrorInvalidName, err)
	})

	t.Run("should return error when score is invalid", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		input := &updategradingitem.Input{
			ID:              1,
			Name:            test.StringPtr("Updated Grading Item"),
			Description:     test.StringPtr("Updated Description"),
			Score:           test.Float64Ptr(-1),
			ScoringStrategy: test.StringPtr("decrease"),
			OutputFormat:    test.StringPtr("text"),
			DependTestcases: &[]int32{3, 4},
		}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, updategradingitem.ErrorInvalidScore, err)
	})

	t.Run("should return error when scoring strategy is empty", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &updategradingitem.Input{
			ID:              1,
			Name:            test.StringPtr("Updated Grading Item"),
			Description:     test.StringPtr("Updated Description"),
			Score:           test.Float64Ptr(20),
			ScoringStrategy: test.StringPtr(""),
			OutputFormat:    test.StringPtr("text"),
			DependTestcases: &[]int32{3, 4},
		}
		mocks.repository.On("FindByID", int(input.ID)).Return(&gradingitem.GradingItem{}, nil).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, updategradingitem.ErrorInvalidScoringStrategy, err)
	})

	t.Run("should return error when output formatter is empty", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &updategradingitem.Input{
			ID:              1,
			Name:            test.StringPtr("Updated Grading Item"),
			Description:     test.StringPtr("Updated Description"),
			Score:           test.Float64Ptr(20),
			ScoringStrategy: test.StringPtr("decrease"),
			OutputFormat:    test.StringPtr(""),
			DependTestcases: &[]int32{3, 4},
		}
		mocks.repository.On("FindByID", int(input.ID)).Return(&gradingitem.GradingItem{}, nil).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, updategradingitem.ErrorInvalidOutputFormat, err)
	})
}
