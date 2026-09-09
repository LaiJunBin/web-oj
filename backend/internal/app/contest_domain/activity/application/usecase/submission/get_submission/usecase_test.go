package getsubmission_test

import (
	getsubmission "judge-backend/internal/app/contest_domain/activity/application/usecase/submission/get_submission"
	"judge-backend/internal/app/contest_domain/activity/application/query/dto"
	"judge-backend/internal/app/contest_domain/activity/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUsecase_Execute(t *testing.T) {
	t.Run("should return submission successfully", func(t *testing.T) {
		q := mocks.NewSubmissionQuery(t)
		uc := getsubmission.NewUsecase(q)

		detail := &dto.SubmissionDetail{ID: "sub-1"}
		q.On("GetSubmissionByID", "sub-1").Return(detail, nil).Once()

		output, err := uc.Execute(&getsubmission.Input{SubmissionID: "sub-1"})

		assert.NoError(t, err)
		assert.Equal(t, detail, output.Submission)
	})

	t.Run("should return error when submission ID is empty", func(t *testing.T) {
		q := mocks.NewSubmissionQuery(t)
		uc := getsubmission.NewUsecase(q)

		_, err := uc.Execute(&getsubmission.Input{SubmissionID: ""})

		assert.ErrorIs(t, err, getsubmission.ErrorRequiredSubmissionID)
	})

	t.Run("should return error when query fails", func(t *testing.T) {
		q := mocks.NewSubmissionQuery(t)
		uc := getsubmission.NewUsecase(q)

		q.On("GetSubmissionByID", "sub-1").Return((*dto.SubmissionDetail)(nil), assert.AnError).Once()

		_, err := uc.Execute(&getsubmission.Input{SubmissionID: "sub-1"})

		assert.Error(t, err)
	})
}
