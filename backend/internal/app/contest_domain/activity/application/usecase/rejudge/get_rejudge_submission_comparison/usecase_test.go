package getrejudgesubmissioncomparison_test

import (
	"errors"
	"testing"

	"judge-backend/internal/app/contest_domain/activity/application/query/dto"
	getrejudgesubmissioncomparison "judge-backend/internal/app/contest_domain/activity/application/usecase/rejudge/get_rejudge_submission_comparison"
	"judge-backend/internal/app/contest_domain/activity/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsecase_Execute(t *testing.T) {
	t.Run("returns the comparison for the requested rejudge and submission", func(t *testing.T) {
		query := mocks.NewRejudgeQuery(t)
		expected := &dto.SubmissionComparisonDetail{
			ID:    "sub-1",
			After: &dto.JudgmentOutcomeDetail{JudgmentID: "judgment-2", Score: 70},
		}
		query.On("GetSubmissionComparison", "rejudge-1", "sub-1").Return(expected, nil)

		output, err := getrejudgesubmissioncomparison.NewUsecase(query).Execute(
			&getrejudgesubmissioncomparison.Input{RejudgeID: "rejudge-1", SubmissionID: "sub-1"},
		)

		require.NoError(t, err)
		require.NotNil(t, output)
		assert.Same(t, expected, output.Comparison)
	})

	t.Run("rejects a missing rejudge id without querying", func(t *testing.T) {
		query := mocks.NewRejudgeQuery(t)

		output, err := getrejudgesubmissioncomparison.NewUsecase(query).Execute(
			&getrejudgesubmissioncomparison.Input{SubmissionID: "sub-1"},
		)

		assert.ErrorIs(t, err, getrejudgesubmissioncomparison.ErrRejudgeIDIsRequired)
		assert.Nil(t, output)
		query.AssertNotCalled(t, "GetSubmissionComparison")
	})

	t.Run("rejects a missing submission id without querying", func(t *testing.T) {
		query := mocks.NewRejudgeQuery(t)

		output, err := getrejudgesubmissioncomparison.NewUsecase(query).Execute(
			&getrejudgesubmissioncomparison.Input{RejudgeID: "rejudge-1"},
		)

		assert.ErrorIs(t, err, getrejudgesubmissioncomparison.ErrSubmissionIDIsRequired)
		assert.Nil(t, output)
		query.AssertNotCalled(t, "GetSubmissionComparison")
	})

	t.Run("propagates the query error unchanged so its status code survives", func(t *testing.T) {
		query := mocks.NewRejudgeQuery(t)
		expected := errors.New("rejudge not found")
		query.On("GetSubmissionComparison", "rejudge-1", "sub-1").Return(nil, expected)

		output, err := getrejudgesubmissioncomparison.NewUsecase(query).Execute(
			&getrejudgesubmissioncomparison.Input{RejudgeID: "rejudge-1", SubmissionID: "sub-1"},
		)

		assert.Same(t, expected, err)
		assert.Nil(t, output)
	})
}
