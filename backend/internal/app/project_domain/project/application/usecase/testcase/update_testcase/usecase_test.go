package updatetestcase_test

import (
	"errors"
	"testing"

	updatetestcase "judge-backend/internal/app/project_domain/project/application/usecase/testcase/update_testcase"
	"judge-backend/internal/app/project_domain/project/domain/entity/testcase"
	"judge-backend/internal/app/project_domain/project/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type UsecaseMocks struct {
	repository *mocks.TestcaseRepository
}

func NewUsecaseMocks(t *testing.T) *UsecaseMocks {
	return &UsecaseMocks{
		repository: mocks.NewTestcaseRepository(t),
	}
}

func NewUsecase(t *testing.T) (updatetestcase.Usecase, *UsecaseMocks) {
	mocks := NewUsecaseMocks(t)
	return updatetestcase.NewUsecase(mocks.repository), mocks
}

func TestExecute(t *testing.T) {
	t.Run("should update testcase successfully", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		name := "Updated Testcase"
		command := "Updated Command"
		score := 10.0
		isSample := true
		depends := []int32{2, 3}
		input := &updatetestcase.Input{
			ID:       1,
			Name:     &name,
			Command:  &command,
			Score:    &score,
			IsSample: &isSample,
			Depends:  &depends,
		}
		testcase := &testcase.Testcase{}
		mocks.repository.On("FindByID", int(input.ID)).Return(testcase, nil).Once()
		mocks.repository.On("Save", mock.AnythingOfType("*testcase.Testcase")).Return(nil).Once()

		output, err := usecase.Execute(input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
	})

	t.Run("should return error when ID is invalid", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		input := &updatetestcase.Input{ID: 0}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, updatetestcase.ErrorInvalidID, err)
	})

	t.Run("should return error when input is empty", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		input := &updatetestcase.Input{ID: 1}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, updatetestcase.ErrorEmptyInput, err)
	})

	t.Run("should return error when score is invalid", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		score := -1.0
		input := &updatetestcase.Input{ID: 1, Score: &score}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, updatetestcase.ErrorInvalidScore, err)
	})

	t.Run("should return error when testcase not found", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		name := "Updated Testcase"
		input := &updatetestcase.Input{ID: 1, Name: &name}
		mocks.repository.On("FindByID", int(input.ID)).Return(nil, nil).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, updatetestcase.ErrorTestcaseNotFound, err)
	})

	t.Run("should return internal server error when repository find fails", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		name := "Updated Testcase"
		input := &updatetestcase.Input{ID: 1, Name: &name}
		mocks.repository.On("FindByID", int(input.ID)).Return(nil, errors.New("find error")).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, updatetestcase.ErrorInternalServer, err)
	})

	t.Run("should return internal server error when repository save fails", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		name := "Updated Testcase"
		input := &updatetestcase.Input{ID: 1, Name: &name}
		testcase := &testcase.Testcase{}
		mocks.repository.On("FindByID", int(input.ID)).Return(testcase, nil).Once()
		mocks.repository.On("Save", mock.AnythingOfType("*testcase.Testcase")).Return(errors.New("save error")).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, updatetestcase.ErrorInternalServer, err)
	})
}
