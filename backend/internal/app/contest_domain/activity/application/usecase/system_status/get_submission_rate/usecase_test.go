package getsubmissionrate_test

import (
	"testing"

	"judge-backend/internal/app/contest_domain/activity/application/query/dto"
	getsubmissionrate "judge-backend/internal/app/contest_domain/activity/application/usecase/system_status/get_submission_rate"
	"judge-backend/internal/app/contest_domain/activity/mocks"

	"github.com/stretchr/testify/assert"
)

func int32Ptr(v int32) *int32 { return &v }

func TestUsecase_Execute(t *testing.T) {
	t.Run("defaults the window when the caller does not ask for one", func(t *testing.T) {
		q := mocks.NewSystemStatusQuery(t)
		uc := getsubmissionrate.NewUsecase(q)

		q.On("GetSubmissionRate", getsubmissionrate.DefaultMinutes, (*int32)(nil)).
			Return([]*dto.SubmissionRatePoint{}, nil).Once()

		_, err := uc.Execute(&getsubmissionrate.Input{})

		assert.NoError(t, err)
	})

	t.Run("passes the requested window and contest project through", func(t *testing.T) {
		q := mocks.NewSystemStatusQuery(t)
		uc := getsubmissionrate.NewUsecase(q)

		contestProjectID := int32Ptr(3)
		q.On("GetSubmissionRate", int32(30), contestProjectID).
			Return([]*dto.SubmissionRatePoint{{Count: 2}}, nil).Once()

		output, err := uc.Execute(&getsubmissionrate.Input{
			Minutes:          int32Ptr(30),
			ContestProjectID: contestProjectID,
		})

		assert.NoError(t, err)
		assert.Len(t, output.Points, 1)
	})

	// The response carries one point per minute whether or not anything
	// happened, so an unbounded window is an unbounded response.
	t.Run("rejects a window outside the supported range without querying", func(t *testing.T) {
		for _, minutes := range []int32{0, -1, getsubmissionrate.MaxMinutes + 1} {
			q := mocks.NewSystemStatusQuery(t)
			uc := getsubmissionrate.NewUsecase(q)

			_, err := uc.Execute(&getsubmissionrate.Input{Minutes: int32Ptr(minutes)})

			assert.ErrorIs(t, err, getsubmissionrate.ErrorInvalidMinutes, "minutes=%d", minutes)
			q.AssertNotCalled(t, "GetSubmissionRate")
		}
	})

	// Widening a malformed id to "the whole system" would answer a question the
	// caller did not ask, and the number would look plausible.
	t.Run("rejects a non-positive contest project ID rather than ignoring it", func(t *testing.T) {
		q := mocks.NewSystemStatusQuery(t)
		uc := getsubmissionrate.NewUsecase(q)

		_, err := uc.Execute(&getsubmissionrate.Input{ContestProjectID: int32Ptr(0)})

		assert.ErrorIs(t, err, getsubmissionrate.ErrorInvalidContestProjectID)
		q.AssertNotCalled(t, "GetSubmissionRate")
	})

	t.Run("returns the error when the query fails", func(t *testing.T) {
		q := mocks.NewSystemStatusQuery(t)
		uc := getsubmissionrate.NewUsecase(q)

		q.On("GetSubmissionRate", getsubmissionrate.DefaultMinutes, (*int32)(nil)).
			Return(([]*dto.SubmissionRatePoint)(nil), assert.AnError).Once()

		_, err := uc.Execute(&getsubmissionrate.Input{})

		assert.Error(t, err)
	})
}
