package rejudgesubmission_test

import (
	rejudgesubmission "judge-backend/internal/app/contest_domain/activity/application/usecase/submission/rejudge_submission"
	"judge-backend/internal/app/contest_domain/activity/domain/entity/rejudge"
	"judge-backend/internal/app/contest_domain/activity/domain/entity/submission"
	"judge-backend/internal/app/contest_domain/activity/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	testifymock "github.com/stretchr/testify/mock"
)

func newSubmission() *submission.Submission {
	return submission.New(&submission.SubmissionConfig{
		ID:               "sub-1",
		UserID:           10,
		ContestProjectID: 1,
	})
}

func TestUsecase_Execute(t *testing.T) {
	t.Run("should rejudge submission successfully", func(t *testing.T) {
		repo := mocks.NewSubmissionRepository(t)
		rejudgeRepo := mocks.NewRejudgeRepository(t)
		svc := mocks.NewSubmissionService(t)
		uc := rejudgesubmission.NewUsecase(repo, rejudgeRepo, svc)

		sub := newSubmission()
		repo.On("FindByID", "sub-1").Return(sub, nil).Once()
		rejudgeRepo.On("FindAwaitingDecisionBySubmissionIDs", []string{"sub-1"}).Return([]*rejudge.Rejudge{}, nil).Once()
		svc.On("RequestRejudgeSubmission", testifymock.Anything, "Test reason").Return().Once()

		output, err := uc.Execute(&rejudgesubmission.Input{
			SubmissionID: "sub-1",
			Reason:       "Test reason",
		})

		assert.NoError(t, err)
		assert.NotNil(t, output)
	})

	t.Run("should return error when submission ID is empty", func(t *testing.T) {
		repo := mocks.NewSubmissionRepository(t)
		rejudgeRepo := mocks.NewRejudgeRepository(t)
		svc := mocks.NewSubmissionService(t)
		uc := rejudgesubmission.NewUsecase(repo, rejudgeRepo, svc)

		_, err := uc.Execute(&rejudgesubmission.Input{
			SubmissionID: "",
			Reason:       "Test reason",
		})

		assert.ErrorIs(t, err, rejudgesubmission.ErrorRequiredSubmissionID)
	})

	t.Run("should return error when reason is empty", func(t *testing.T) {
		repo := mocks.NewSubmissionRepository(t)
		rejudgeRepo := mocks.NewRejudgeRepository(t)
		svc := mocks.NewSubmissionService(t)
		uc := rejudgesubmission.NewUsecase(repo, rejudgeRepo, svc)

		_, err := uc.Execute(&rejudgesubmission.Input{
			SubmissionID: "sub-1",
			Reason:       "",
		})

		assert.ErrorIs(t, err, rejudgesubmission.ErrorRequiredReason)
	})

	t.Run("should return error when FindByID fails", func(t *testing.T) {
		repo := mocks.NewSubmissionRepository(t)
		rejudgeRepo := mocks.NewRejudgeRepository(t)
		svc := mocks.NewSubmissionService(t)
		uc := rejudgesubmission.NewUsecase(repo, rejudgeRepo, svc)

		repo.On("FindByID", "sub-1").Return((*submission.Submission)(nil), assert.AnError).Once()

		_, err := uc.Execute(&rejudgesubmission.Input{
			SubmissionID: "sub-1",
			Reason:       "Test reason",
		})

		assert.Error(t, err)
	})

	t.Run("refuses a second rejudge while one is still awaiting a decision", func(t *testing.T) {
		repo := mocks.NewSubmissionRepository(t)
		rejudgeRepo := mocks.NewRejudgeRepository(t)
		svc := mocks.NewSubmissionService(t)
		uc := rejudgesubmission.NewUsecase(repo, rejudgeRepo, svc)

		outstanding, err := rejudge.New(&rejudge.RejudgeConfig{ID: "rejudge-1", Reason: "earlier", Status: "Finished"})
		assert.NoError(t, err)

		repo.On("FindByID", "sub-1").Return(newSubmission(), nil).Once()
		rejudgeRepo.On("FindAwaitingDecisionBySubmissionIDs", []string{"sub-1"}).
			Return([]*rejudge.Rejudge{outstanding}, nil).Once()

		_, err = uc.Execute(&rejudgesubmission.Input{
			SubmissionID: "sub-1",
			Reason:       "Test reason",
		})

		assert.ErrorIs(t, err, rejudgesubmission.ErrorRejudgeAlreadyAwaitingDecision)
		svc.AssertNotCalled(t, "RequestRejudgeSubmission")
	})

	t.Run("propagates a failure to look up outstanding rejudges", func(t *testing.T) {
		repo := mocks.NewSubmissionRepository(t)
		rejudgeRepo := mocks.NewRejudgeRepository(t)
		svc := mocks.NewSubmissionService(t)
		uc := rejudgesubmission.NewUsecase(repo, rejudgeRepo, svc)

		repo.On("FindByID", "sub-1").Return(newSubmission(), nil).Once()
		rejudgeRepo.On("FindAwaitingDecisionBySubmissionIDs", []string{"sub-1"}).
			Return(nil, assert.AnError).Once()

		_, err := uc.Execute(&rejudgesubmission.Input{
			SubmissionID: "sub-1",
			Reason:       "Test reason",
		})

		assert.Error(t, err)
		svc.AssertNotCalled(t, "RequestRejudgeSubmission")
	})
}
