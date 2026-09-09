package approverejudge_test

import (
	approverejudge "judge-backend/internal/app/contest_domain/activity/application/usecase/rejudge/approve_rejudge"
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
	t.Run("should approve rejudge successfully", func(t *testing.T) {
		repo := mocks.NewRejudgeRepository(t)
		svc := mocks.NewRejudgeService(t)
		uc := approverejudge.NewUsecase(repo, svc)

		r := newRejudge(t, "Finished")
		repo.On("FindByID", "rejudge-1").Return(r, nil).Once()
		svc.On("RequestApproveRejudge", testifymock.Anything).Return().Once()

		output, err := uc.Execute(&approverejudge.Input{RejudgeID: "rejudge-1"})

		assert.NoError(t, err)
		assert.NotNil(t, output)
	})

	t.Run("should return error when rejudge ID is empty", func(t *testing.T) {
		repo := mocks.NewRejudgeRepository(t)
		svc := mocks.NewRejudgeService(t)
		uc := approverejudge.NewUsecase(repo, svc)

		_, err := uc.Execute(&approverejudge.Input{RejudgeID: ""})

		assert.ErrorIs(t, err, approverejudge.ErrorRejudgeIDIsRequired)
	})

	t.Run("should return error when FindByID fails", func(t *testing.T) {
		repo := mocks.NewRejudgeRepository(t)
		svc := mocks.NewRejudgeService(t)
		uc := approverejudge.NewUsecase(repo, svc)

		repo.On("FindByID", "rejudge-1").Return((*rejudge.Rejudge)(nil), assert.AnError).Once()

		_, err := uc.Execute(&approverejudge.Input{RejudgeID: "rejudge-1"})

		assert.Error(t, err)
	})

	t.Run("should return error when rejudge cannot be approved", func(t *testing.T) {
		repo := mocks.NewRejudgeRepository(t)
		svc := mocks.NewRejudgeService(t)
		uc := approverejudge.NewUsecase(repo, svc)

		// Pending status cannot be approved (only Finished can)
		r := newRejudge(t, "Pending")
		repo.On("FindByID", "rejudge-1").Return(r, nil).Once()

		_, err := uc.Execute(&approverejudge.Input{RejudgeID: "rejudge-1"})

		assert.ErrorIs(t, err, approverejudge.ErrorRejudgeCannotBeApproved)
	})
}
