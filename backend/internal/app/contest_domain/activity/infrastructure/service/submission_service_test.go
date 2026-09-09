package service

import (
	"testing"

	"judge-backend/internal/app/contest_domain/activity/domain/entity/submission"
	sharedevent "judge-backend/internal/app/shared/events"
	sharedmocks "judge-backend/internal/app/shared/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// The rejudge record is grouped by contest project in the admin UI, and this is
// the only place the project is attached: nothing downstream of the event can
// recover it, because the judge domain is handed submission ids and nothing
// else. A request published without it produces a rejudge belonging to project
// 0.
func TestSubmissionService_RequestRejudge_CarriesTheContestProject(t *testing.T) {
	t.Run("a single submission contributes its own contest project", func(t *testing.T) {
		bus := sharedmocks.NewEventBus(t)
		svc := NewSubmissionService(bus)

		var published *sharedevent.RejudgeRequested
		bus.On("Publish", mock.Anything).Run(func(args mock.Arguments) {
			published = args.Get(0).(*sharedevent.RejudgeRequested)
		}).Once()

		svc.RequestRejudgeSubmission(
			submission.New(&submission.SubmissionConfig{ID: "sub-1", ContestProjectID: 7}),
			"reason",
		)

		require.NotNil(t, published)
		assert.EqualValues(t, 7, published.ContestProjectID())
		assert.Equal(t, []string{"sub-1"}, published.SubmissionIDs())
	})

	t.Run("a whole-project rejudge uses the project it was asked for", func(t *testing.T) {
		bus := sharedmocks.NewEventBus(t)
		svc := NewSubmissionService(bus)

		var published *sharedevent.RejudgeRequested
		bus.On("Publish", mock.Anything).Run(func(args mock.Arguments) {
			published = args.Get(0).(*sharedevent.RejudgeRequested)
		}).Once()

		svc.RequestRejudgeSubmissions(7, []*submission.Submission{
			submission.New(&submission.SubmissionConfig{ID: "sub-1", ContestProjectID: 7}),
			submission.New(&submission.SubmissionConfig{ID: "sub-2", ContestProjectID: 7}),
		}, "reason")

		require.NotNil(t, published)
		assert.EqualValues(t, 7, published.ContestProjectID())
		assert.Equal(t, []string{"sub-1", "sub-2"}, published.SubmissionIDs())
	})

	// Nothing in the submission list can supply the project here, which is why
	// the caller passes it separately.
	t.Run("a project with no submissions still names its project", func(t *testing.T) {
		bus := sharedmocks.NewEventBus(t)
		svc := NewSubmissionService(bus)

		var published *sharedevent.RejudgeRequested
		bus.On("Publish", mock.Anything).Run(func(args mock.Arguments) {
			published = args.Get(0).(*sharedevent.RejudgeRequested)
		}).Once()

		svc.RequestRejudgeSubmissions(7, []*submission.Submission{}, "reason")

		require.NotNil(t, published)
		assert.EqualValues(t, 7, published.ContestProjectID())
		assert.Empty(t, published.SubmissionIDs())
	})
}
