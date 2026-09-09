package deletetestcase_test

import (
	"errors"
	"testing"

	deletetestcase "judge-backend/internal/app/project_domain/project/application/usecase/testcase/delete_testcase"
	"judge-backend/internal/app/project_domain/project/domain/entity/testcase"
	"judge-backend/internal/app/project_domain/project/mocks"

	"github.com/stretchr/testify/assert"
)

type UsecaseMocks struct {
	repository *mocks.TestcaseRepository
}

func NewUsecaseMocks(t *testing.T) *UsecaseMocks {
	return &UsecaseMocks{
		repository: mocks.NewTestcaseRepository(t),
	}
}

func NewUsecase(t *testing.T) (deletetestcase.Usecase, *UsecaseMocks) {
	mocks := NewUsecaseMocks(t)
	return deletetestcase.NewUsecase(mocks.repository), mocks
}

func TestExecute(t *testing.T) {
	t.Run("should delete testcase successfully", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &deletetestcase.Input{ID: 1}
		testcase := &testcase.Testcase{}
		mocks.repository.On("FindByID", int(input.ID)).Return(testcase, nil).Once()
		mocks.repository.On("Delete", testcase).Return(nil).Once()

		output, err := usecase.Execute(input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
	})

	t.Run("should return error when ID is invalid", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		input := &deletetestcase.Input{ID: 0}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, deletetestcase.ErrorInvalidID, err)
	})

	t.Run("should return error when testcase not found", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &deletetestcase.Input{ID: 1}
		mocks.repository.On("FindByID", int(input.ID)).Return(nil, nil).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, deletetestcase.ErrorTestcaseNotFound, err)
	})

	t.Run("should return internal server error when repository find fails", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &deletetestcase.Input{ID: 1}
		mocks.repository.On("FindByID", int(input.ID)).Return(nil, errors.New("find error")).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, deletetestcase.ErrorInternalServer, err)
	})

	t.Run("should return internal server error when repository delete fails", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &deletetestcase.Input{ID: 1}
		testcase := &testcase.Testcase{}
		mocks.repository.On("FindByID", int(input.ID)).Return(testcase, nil).Once()
		mocks.repository.On("Delete", testcase).Return(errors.New("delete error")).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, deletetestcase.ErrorInternalServer, err)
	})
}
