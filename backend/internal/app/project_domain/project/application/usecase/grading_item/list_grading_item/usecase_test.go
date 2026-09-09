package listgradingitem_test

import (
	"errors"
	"testing"

	"judge-backend/internal/app/project_domain/project/application/query/dto"
	listgradingitem "judge-backend/internal/app/project_domain/project/application/usecase/grading_item/list_grading_item"
	"judge-backend/internal/app/project_domain/project/mocks"

	"github.com/stretchr/testify/assert"
)

type UsecaseMocks struct {
	query       *mocks.GradingItemQuery
	projectRepo *mocks.ProjectRepository
}

func NewUsecaseMocks(t *testing.T) *UsecaseMocks {
	return &UsecaseMocks{
		query:       mocks.NewGradingItemQuery(t),
		projectRepo: mocks.NewProjectRepository(t),
	}
}

func NewUsecase(t *testing.T) (listgradingitem.Usecase, *UsecaseMocks) {
	mocks := NewUsecaseMocks(t)
	return listgradingitem.NewUsecase(mocks.query, mocks.projectRepo), mocks
}

func TestExecute(t *testing.T) {
	t.Run("should list grading items successfully", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &listgradingitem.Input{ProjectID: 1}
		gradingItems := []*dto.GradingItem{}
		mocks.projectRepo.On("Exists", int(input.ProjectID)).Return(true, nil).Once()
		mocks.query.On("ListGradingItem", input.ProjectID).Return(gradingItems, nil).Once()

		output, err := usecase.Execute(input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, gradingItems, output.GradingItems)
	})

	t.Run("should return error when project ID is invalid", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		input := &listgradingitem.Input{ProjectID: 0}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, listgradingitem.ErrorInvalidProjectID, err)
	})

	t.Run("should return error when project not found", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &listgradingitem.Input{ProjectID: 1}
		mocks.projectRepo.On("Exists", int(input.ProjectID)).Return(false, nil).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, listgradingitem.ErrorProjectNotFound, err)
	})

	t.Run("should return internal server error when project repo fails", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &listgradingitem.Input{ProjectID: 1}
		mocks.projectRepo.On("Exists", int(input.ProjectID)).Return(false, errors.New("repo error")).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, listgradingitem.ErrorInternalServer, err)
	})

	t.Run("should return internal server error when query fails", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &listgradingitem.Input{ProjectID: 1}
		mocks.projectRepo.On("Exists", int(input.ProjectID)).Return(true, nil).Once()
		mocks.query.On("ListGradingItem", input.ProjectID).Return(nil, errors.New("query error")).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, listgradingitem.ErrorInternalServer, err)
	})
}
