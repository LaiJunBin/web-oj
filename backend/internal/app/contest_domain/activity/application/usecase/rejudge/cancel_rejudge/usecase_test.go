package cancelrejudge_test

import (
	cancelrejudge "judge-backend/internal/app/contest_domain/activity/application/usecase/rejudge/cancel_rejudge"
	"judge-backend/internal/app/contest_domain/activity/domain/entity/rejudge"
	"judge-backend/internal/app/contest_domain/activity/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	testifymock "github.com/stretchr/testify/mock"
)

func newRejudge(t *testing.T, status string) *rejudge.Rejudge {
	t.Helper()
	r, err := rejudge.New(&rejudge.RejudgeConfig{
		ID:     "rejudge-1",
		Reason: "Test reason",
		Status: status,
	})
	if err != nil {
		t.Fatalf("failed to create rejudge: %v", err)
	}
	return r
}

func TestUsecase_Execute(t *testing.T) {
	t.Run("should cancel pending rejudge successfully", func(t *testing.T) {
		repo := mocks.NewRejudgeRepository(t)
		svc := mocks.NewRejudgeService(t)
		uc := cancelrejudge.NewUsecase(repo, svc)

		r := newRejudge(t, "Pending")
		repo.On("FindByID", "rejudge-1").Return(r, nil).Once()
		svc.On("RequestCancelRejudge", testifymock.Anything).Return().Once()

		output, err := uc.Execute(&cancelrejudge.Input{RejudgeID: "rejudge-1"})

		assert.NoError(t, err)
		assert.NotNil(t, output)
	})

	t.Run("should cancel running rejudge successfully", func(t *testing.T) {
		repo := mocks.NewRejudgeRepository(t)
		svc := mocks.NewRejudgeService(t)
		uc := cancelrejudge.NewUsecase(repo, svc)

		r := newRejudge(t, "Running")
		repo.On("FindByID", "rejudge-1").Return(r, nil).Once()
		svc.On("RequestCancelRejudge", testifymock.Anything).Return().Once()

		output, err := uc.Execute(&cancelrejudge.Input{RejudgeID: "rejudge-1"})

		assert.NoError(t, err)
		assert.NotNil(t, output)
	})

	t.Run("should return error when rejudge ID is empty", func(t *testing.T) {
		repo := mocks.NewRejudgeRepository(t)
		svc := mocks.NewRejudgeService(t)
		uc := cancelrejudge.NewUsecase(repo, svc)

		_, err := uc.Execute(&cancelrejudge.Input{RejudgeID: ""})

		assert.ErrorIs(t, err, cancelrejudge.ErrorRejudgeIDIsRequired)
	})

	t.Run("should return error when FindByID fails", func(t *testing.T) {
		repo := mocks.NewRejudgeRepository(t)
		svc := mocks.NewRejudgeService(t)
		uc := cancelrejudge.NewUsecase(repo, svc)

		repo.On("FindByID", "rejudge-1").Return((*rejudge.Rejudge)(nil), assert.AnError).Once()

		_, err := uc.Execute(&cancelrejudge.Input{RejudgeID: "rejudge-1"})

		assert.Error(t, err)
	})

	t.Run("should return error when rejudge cannot be cancelled", func(t *testing.T) {
		repo := mocks.NewRejudgeRepository(t)
		svc := mocks.NewRejudgeService(t)
		uc := cancelrejudge.NewUsecase(repo, svc)

		// Finished status cannot be cancelled
		r := newRejudge(t, "Finished")
		repo.On("FindByID", "rejudge-1").Return(r, nil).Once()

		_, err := uc.Execute(&cancelrejudge.Input{RejudgeID: "rejudge-1"})

		assert.ErrorIs(t, err, cancelrejudge.ErrorJudgmentCannotBeCancelled)
	})

	// A CANCELED rejudge accepts a repeated cancel request as an idempotent
	// retry: a previous cancellation can have flipped the parent and then
	// failed partway through its members, and the republished event is what
	// lets the judge-domain handler reconcile those stragglers. Rejecting
	// here made that recovery path unreachable — no user action could ever
	// re-trigger it.
	t.Run("should accept a repeated cancel of an already canceled rejudge and republish the event", func(t *testing.T) {
		repo := mocks.NewRejudgeRepository(t)
		svc := mocks.NewRejudgeService(t)
		uc := cancelrejudge.NewUsecase(repo, svc)

		r := newRejudge(t, "Canceled")
		repo.On("FindByID", "rejudge-1").Return(r, nil).Once()
		svc.On("RequestCancelRejudge", r).Return().Once()

		output, err := uc.Execute(&cancelrejudge.Input{RejudgeID: "rejudge-1"})

		assert.NoError(t, err)
		assert.NotNil(t, output)
	})

	// The idempotent-retry carve-out is for CANCELED only — every other
	// terminal status keeps rejecting.
	t.Run("should still reject other terminal statuses", func(t *testing.T) {
		for _, status := range []string{"Approved", "Rejected", "Error"} {
			repo := mocks.NewRejudgeRepository(t)
			svc := mocks.NewRejudgeService(t)
			uc := cancelrejudge.NewUsecase(repo, svc)

			r := newRejudge(t, status)
			repo.On("FindByID", "rejudge-1").Return(r, nil).Once()
			// No RequestCancelRejudge expectation: an unexpected call panics.

			_, err := uc.Execute(&cancelrejudge.Input{RejudgeID: "rejudge-1"})

			assert.ErrorIs(t, err, cancelrejudge.ErrorJudgmentCannotBeCancelled, "status %s must stay rejected", status)
		}
	})
}
