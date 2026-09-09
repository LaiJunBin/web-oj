package getsubmissionsbycontestproject_test

import (
	getsubmissionsbycontestproject "judge-backend/internal/app/contest_domain/activity/application/usecase/submission/get_submissions_by_contest_project"
	"judge-backend/internal/app/contest_domain/activity/application/query/dto"
	"judge-backend/internal/app/contest_domain/activity/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUsecase_Execute(t *testing.T) {
	t.Run("should return submissions successfully", func(t *testing.T) {
		q := mocks.NewSubmissionQuery(t)
		uc := getsubmissionsbycontestproject.NewUsecase(q)

		submissions := []*dto.Submission{
			{ID: "sub-1"},
			{ID: "sub-2"},
		}
		q.On("GetSubmissionsByContestProjectID", int32(1)).Return(submissions, nil).Once()

		output, err := uc.Execute(&getsubmissionsbycontestproject.Input{ContestProjectID: 1})

		assert.NoError(t, err)
		assert.Len(t, output.Submissions, 2)
	})

	t.Run("should return error when contest project ID is invalid", func(t *testing.T) {
		q := mocks.NewSubmissionQuery(t)
		uc := getsubmissionsbycontestproject.NewUsecase(q)

		_, err := uc.Execute(&getsubmissionsbycontestproject.Input{ContestProjectID: 0})

		assert.ErrorIs(t, err, getsubmissionsbycontestproject.ErrorInvalidContestProjectID)
	})

	t.Run("should return error when query fails", func(t *testing.T) {
		q := mocks.NewSubmissionQuery(t)
		uc := getsubmissionsbycontestproject.NewUsecase(q)

		q.On("GetSubmissionsByContestProjectID", int32(1)).Return(([]*dto.Submission)(nil), assert.AnError).Once()

		_, err := uc.Execute(&getsubmissionsbycontestproject.Input{ContestProjectID: 1})

		assert.Error(t, err)
	})
}
