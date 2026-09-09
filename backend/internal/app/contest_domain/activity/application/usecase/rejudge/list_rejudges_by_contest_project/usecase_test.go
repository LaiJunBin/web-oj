package listrejudgesbycontestproject_test

import (
	"judge-backend/internal/app/contest_domain/activity/application/query/dto"
	listrejudgesbycontestproject "judge-backend/internal/app/contest_domain/activity/application/usecase/rejudge/list_rejudges_by_contest_project"
	"judge-backend/internal/app/contest_domain/activity/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUsecase_Execute(t *testing.T) {
	t.Run("should list the contest project's rejudges", func(t *testing.T) {
		q := mocks.NewRejudgeQuery(t)
		uc := listrejudgesbycontestproject.NewUsecase(q)

		rejudges := []*dto.RejudgeInfo{
			{ID: "rejudge-1", ContestProjectID: 3, Status: "Finished"},
			{ID: "rejudge-2", ContestProjectID: 3, Status: "Pending"},
		}
		q.On("ListRejudgesByContestProjectID", int32(3)).Return(rejudges, nil).Once()

		output, err := uc.Execute(&listrejudgesbycontestproject.Input{ContestProjectID: 3})

		assert.NoError(t, err)
		assert.Len(t, output.Rejudges, 2)
	})

	t.Run("should return empty list when the project has no rejudges", func(t *testing.T) {
		q := mocks.NewRejudgeQuery(t)
		uc := listrejudgesbycontestproject.NewUsecase(q)

		q.On("ListRejudgesByContestProjectID", int32(3)).Return([]*dto.RejudgeInfo{}, nil).Once()

		output, err := uc.Execute(&listrejudgesbycontestproject.Input{ContestProjectID: 3})

		assert.NoError(t, err)
		assert.Empty(t, output.Rejudges)
	})

	// Rejudges are stored with a positive contest project id, so a
	// non-positive one can only be a malformed request — answering it with an
	// empty list would read as "this project has no rejudges".
	t.Run("should reject a non-positive contest project ID without querying", func(t *testing.T) {
		q := mocks.NewRejudgeQuery(t)
		uc := listrejudgesbycontestproject.NewUsecase(q)

		_, err := uc.Execute(&listrejudgesbycontestproject.Input{ContestProjectID: 0})

		assert.ErrorIs(t, err, listrejudgesbycontestproject.ErrorInvalidContestProjectID)
		q.AssertNotCalled(t, "ListRejudgesByContestProjectID")
	})

	t.Run("should return error when query fails", func(t *testing.T) {
		q := mocks.NewRejudgeQuery(t)
		uc := listrejudgesbycontestproject.NewUsecase(q)

		q.On("ListRejudgesByContestProjectID", int32(3)).Return(([]*dto.RejudgeInfo)(nil), assert.AnError).Once()

		_, err := uc.Execute(&listrejudgesbycontestproject.Input{ContestProjectID: 3})

		assert.Error(t, err)
	})
}
