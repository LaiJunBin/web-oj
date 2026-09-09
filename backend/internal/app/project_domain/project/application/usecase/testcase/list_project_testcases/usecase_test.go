package listprojecttestcases_test

import (
	"errors"
	"testing"

	listprojecttestcases "judge-backend/internal/app/project_domain/project/application/usecase/testcase/list_project_testcases"
	"judge-backend/internal/app/project_domain/project/application/query/dto"
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

func NewUsecase(t *testing.T) (listprojecttestcases.Usecase, *UsecaseMocks) {
	m := NewUsecaseMocks(t)
	return listprojecttestcases.NewUsecase(m.query), m
}

func TestExecute(t *testing.T) {
	t.Run("should return list of testcases successfully", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		input := &listprojecttestcases.Input{ProjectID: 1}
		gradingItems := []*dto.ProjectTestcaseGradingItem{
			{ID: 1, Name: "GradingItem1", Testcases: []*dto.ProjectTestcaseInfo{{ID: 1, Name: "TC1"}}},
		}
		m.query.On("ListProjectTestcases", input.ProjectID).Return(gradingItems, nil).Once()

		output, err := usecase.Execute(input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, gradingItems, output.GradingItems)
	})

	t.Run("should return empty list when no testcases", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		input := &listprojecttestcases.Input{ProjectID: 1}
		m.query.On("ListProjectTestcases", input.ProjectID).Return([]*dto.ProjectTestcaseGradingItem{}, nil).Once()

		output, err := usecase.Execute(input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Empty(t, output.GradingItems)
	})

	t.Run("should return error when project ID is invalid", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		input := &listprojecttestcases.Input{ProjectID: 0}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, listprojecttestcases.ErrorInvalidProjectID, err)
	})

	t.Run("should return error when query fails", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		input := &listprojecttestcases.Input{ProjectID: 1}
		queryErr := errors.New("query error")
		m.query.On("ListProjectTestcases", input.ProjectID).Return(nil, queryErr).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, queryErr, err)
	})
}
