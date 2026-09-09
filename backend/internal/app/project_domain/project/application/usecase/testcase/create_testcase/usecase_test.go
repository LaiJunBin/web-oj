package createtestcase_test

import (
	"errors"
	"testing"

	createtestcase "judge-backend/internal/app/project_domain/project/application/usecase/testcase/create_testcase"
	"judge-backend/internal/app/project_domain/project/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type UsecaseMocks struct {
	repository      *mocks.TestcaseRepository
	gradingItemRepo *mocks.GradingItemRepository
}

func NewUsecaseMocks(t *testing.T) *UsecaseMocks {
	return &UsecaseMocks{
		repository:      mocks.NewTestcaseRepository(t),
		gradingItemRepo: mocks.NewGradingItemRepository(t),
	}
}

func NewUsecase(t *testing.T) (createtestcase.Usecase, *UsecaseMocks) {
	mocks := NewUsecaseMocks(t)
	return createtestcase.NewUsecase(mocks.repository, mocks.gradingItemRepo), mocks
}

func TestExecute(t *testing.T) {
	t.Run("should create testcase successfully", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &createtestcase.Input{
			GradingItemID: 1,
			Name:          "Test Testcase",
			Command:       "Command",
			Score:         10,
			IsSample:      true,
			Depends:       []int32{1, 2},
		}
		mocks.gradingItemRepo.On("Exists", int(input.GradingItemID)).Return(true, nil).Once()
		mocks.repository.On("Save", mock.AnythingOfType("*testcase.Testcase")).Return(nil).Once()

		output, err := usecase.Execute(input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
	})

	t.Run("should return error when grading item ID is invalid", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		input := &createtestcase.Input{GradingItemID: 0}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, createtestcase.ErrorInvalidGradingItemID, err)
	})

	t.Run("should return error when grading item not found", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &createtestcase.Input{
			GradingItemID: 1,
			Name:          "Test Testcase",
			Command:       "Command",
			Score:         10,
			IsSample:      true,
			Depends:       []int32{1, 2},
		}
		mocks.gradingItemRepo.On("Exists", int(input.GradingItemID)).Return(false, nil).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, createtestcase.ErrorGradingItemNotFound, err)
	})

	t.Run("should return internal server error when project repo fails", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &createtestcase.Input{
			GradingItemID: 1,
			Name:          "Test Testcase",
			Command:       "Command",
			Score:         10,
			IsSample:      true,
			Depends:       []int32{1, 2},
		}
		mocks.gradingItemRepo.On("Exists", int(input.GradingItemID)).Return(false, errors.New("repo error")).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, createtestcase.ErrorInternalServer, err)
	})

	t.Run("should return error when name is empty", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		input := &createtestcase.Input{
			GradingItemID: 1,
			Name:          "",
			Command:       "Command",
			Score:         10,
			IsSample:      true,
			Depends:       []int32{1, 2},
		}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, createtestcase.ErrorEmptyName, err)
	})

	t.Run("should return error when command is empty", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		input := &createtestcase.Input{
			GradingItemID: 1,
			Name:          "Test Testcase",
			Command:       "",
			Score:         10,
			IsSample:      true,
			Depends:       []int32{1, 2},
		}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, createtestcase.ErrorEmptyCommand, err)
	})

	t.Run("should return error when score is invalid", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		input := &createtestcase.Input{
			GradingItemID: 1,
			Name:          "Test Testcase",
			Command:       "Command",
			Score:         -1,
			IsSample:      true,
			Depends:       []int32{1, 2},
		}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, createtestcase.ErrorInvalidScore, err)
	})

	t.Run("should return error when testcase save fails", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &createtestcase.Input{
			GradingItemID: 1,
			Name:          "Test Testcase",
			Command:       "Command",
			Score:         10,
			IsSample:      true,
			Depends:       []int32{1, 2},
		}
		mocks.gradingItemRepo.On("Exists", int(input.GradingItemID)).Return(true, nil).Once()
		mocks.repository.On("Save", mock.AnythingOfType("*testcase.Testcase")).Return(errors.New("save error")).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, createtestcase.ErrorInternalServer, err)
	})
}
