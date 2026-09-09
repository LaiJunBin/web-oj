package getcontestproject_test

import (
	getcontestproject "judge-backend/internal/app/contest_domain/activity/application/usecase/contest_project/get_contest_project"
	"judge-backend/internal/app/contest_domain/activity/application/query/dto"
	"judge-backend/internal/app/contest_domain/activity/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUsecase_Execute(t *testing.T) {
	t.Run("should return contest project successfully", func(t *testing.T) {
		q := mocks.NewContestProjectQuery(t)
		uc := getcontestproject.NewUsecase(q)

		project := &dto.ContestProject{ID: 1, Name: "Test Project"}
		q.On("GetContestProjectByID", int32(1)).Return(project, nil).Once()

		output, err := uc.Execute(&getcontestproject.Input{ID: 1})

		assert.NoError(t, err)
		assert.Equal(t, project, output.Project)
	})

	t.Run("should return error when query fails", func(t *testing.T) {
		q := mocks.NewContestProjectQuery(t)
		uc := getcontestproject.NewUsecase(q)

		q.On("GetContestProjectByID", int32(1)).Return((*dto.ContestProject)(nil), assert.AnError).Once()

		_, err := uc.Execute(&getcontestproject.Input{ID: 1})

		assert.Error(t, err)
	})
}
