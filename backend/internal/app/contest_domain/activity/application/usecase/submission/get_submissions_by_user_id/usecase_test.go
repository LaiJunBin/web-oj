package getsubmissionsbyuserid_test

import (
	getsubmissionsbyuserid "judge-backend/internal/app/contest_domain/activity/application/usecase/submission/get_submissions_by_user_id"
	"judge-backend/internal/app/contest_domain/activity/application/query/dto"
	"judge-backend/internal/app/contest_domain/activity/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUsecase_Execute(t *testing.T) {
	t.Run("should return submissions by user successfully", func(t *testing.T) {
		q := mocks.NewSubmissionQuery(t)
		uc := getsubmissionsbyuserid.NewUsecase(q)

		submissions := []*dto.Submission{
			{ID: "sub-1"},
		}
		q.On("GetSubmissionsByUserID", int32(10)).Return(submissions, nil).Once()

		output, err := uc.Execute(&getsubmissionsbyuserid.Input{UserID: 10})

		assert.NoError(t, err)
		assert.Len(t, output.Submissions, 1)
	})

	t.Run("should return error when user ID is invalid", func(t *testing.T) {
		q := mocks.NewSubmissionQuery(t)
		uc := getsubmissionsbyuserid.NewUsecase(q)

		_, err := uc.Execute(&getsubmissionsbyuserid.Input{UserID: 0})

		assert.ErrorIs(t, err, getsubmissionsbyuserid.ErrorInvalidUserID)
	})

	t.Run("should return error when query fails", func(t *testing.T) {
		q := mocks.NewSubmissionQuery(t)
		uc := getsubmissionsbyuserid.NewUsecase(q)

		q.On("GetSubmissionsByUserID", int32(10)).Return(([]*dto.Submission)(nil), assert.AnError).Once()

		_, err := uc.Execute(&getsubmissionsbyuserid.Input{UserID: 10})

		assert.Error(t, err)
	})
}
