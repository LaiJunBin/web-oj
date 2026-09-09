package streamusersubmission_test

import (
	"testing"

	"judge-backend/internal/app/contest_domain/contest/application/query/dto"
	streamusersubmission "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/stream_user_submission"
	"judge-backend/internal/app/contest_domain/contest/mocks"
	sharedlogger "judge-backend/internal/app/shared/infrastructure/logger"

	"github.com/stretchr/testify/assert"
)

func TestUsecase_Execute(t *testing.T) {
	t.Run("seeds the stream's resolver with the problem's sample set", func(t *testing.T) {
		q := mocks.NewSubmissionQuery(t)
		q.On("GetUserSubmissionByID", int32(1), "sub-1").Return(&dto.SubmissionDetail{
			SampleTestcaseIDs: []int32{1, 3},
			ContestProjectID:  10,
		}, nil).Once()
		// Within the TTL the resolver serves the seed — no re-query, which is
		// why no GetSampleTestcaseIDs expectation is registered.

		u := streamusersubmission.NewUsecase(q, sharedlogger.NewMockLogger())
		out, err := u.Execute(&streamusersubmission.Input{UserID: 1, SubmissionID: "sub-1"})

		assert.NoError(t, err)
		assert.Equal(t, "sub-1", out.SubmissionID)
		assert.Equal(t, map[int32]bool{1: true, 3: true}, out.ResolveSampleTestcaseIDs())
	})

	t.Run("a missing submission keeps its 404", func(t *testing.T) {
		q := mocks.NewSubmissionQuery(t)
		q.On("GetUserSubmissionByID", int32(1), "sub-x").Return(nil, nil).Once()

		u := streamusersubmission.NewUsecase(q, sharedlogger.NewMockLogger())
		_, err := u.Execute(&streamusersubmission.Input{UserID: 1, SubmissionID: "sub-x"})

		assert.ErrorIs(t, err, streamusersubmission.ErrorSubmissionNotFound)
	})
}
