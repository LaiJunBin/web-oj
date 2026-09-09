package gettestcase_test

import (
	"errors"
	"testing"

	"judge-backend/internal/app/project_domain/project/application/query/dto"
	gettestcase "judge-backend/internal/app/project_domain/project/application/usecase/testcase/get_testcase"
	"judge-backend/internal/app/project_domain/project/mocks"

	"github.com/stretchr/testify/assert"
)

type UsecaseMocks struct {
	query *mocks.TestcaseQuery
}

func NewUsecaseMocks(t *testing.T) *UsecaseMocks {
	return &UsecaseMocks{
		query: mocks.NewTestcaseQuery(t),
	}
}

func NewUsecase(t *testing.T) (gettestcase.Usecase, *UsecaseMocks) {
	mocks := NewUsecaseMocks(t)
	return gettestcase.NewUsecase(mocks.query), mocks
}

func TestExecute(t *testing.T) {
	t.Run("should get testcase successfully", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &gettestcase.Input{ID: 1}
		testcase := &dto.TestcaseDetail{}
		mocks.query.On("GetTestcase", input.ID).Return(testcase, nil).Once()

		output, err := usecase.Execute(input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, testcase, output.Testcase)
	})

	t.Run("should return error when ID is invalid", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		input := &gettestcase.Input{ID: 0}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, gettestcase.ErrorInvalidID, err)
	})

	t.Run("should return error when testcase not found", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &gettestcase.Input{ID: 1}
		mocks.query.On("GetTestcase", input.ID).Return(nil, nil).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, gettestcase.ErrorTestcaseNotFound, err)
	})

	t.Run("should return internal server error when query fails", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &gettestcase.Input{ID: 1}
		mocks.query.On("GetTestcase", input.ID).Return(nil, errors.New("query error")).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, gettestcase.ErrorInternalServer, err)
	})
}
