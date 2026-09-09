package rejudgecontestproject_test

import (
	rejudgecontestproject "judge-backend/internal/app/contest_domain/activity/application/usecase/contest_project/rejudge_contest_project"
	"judge-backend/internal/app/contest_domain/activity/domain/entity/rejudge"
	"judge-backend/internal/app/contest_domain/activity/domain/entity/submission"
	"judge-backend/internal/app/contest_domain/activity/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	testifymock "github.com/stretchr/testify/mock"
)

func validInput() *rejudgecontestproject.Input {
	return &rejudgecontestproject.Input{
		ContestProjectID: 1,
		Reason:           "Test reason",
	}
}

func TestUsecase_Execute(t *testing.T) {
	t.Run("should rejudge contest project successfully", func(t *testing.T) {
		repo := mocks.NewSubmissionRepository(t)
		rejudgeRepo := mocks.NewRejudgeRepository(t)
		svc := mocks.NewSubmissionService(t)
		uc := rejudgecontestproject.NewUsecase(repo, rejudgeRepo, svc)

		submissions := []*submission.Submission{
			submission.New(&submission.SubmissionConfig{ID: "sub-1", ContestProjectID: 1}),
		}
		repo.On("FindByContestProjectIDAndGradingItemIDs", int32(1), []int32{}).Return(submissions, nil).Once()
		rejudgeRepo.On("FindAwaitingDecisionBySubmissionIDs", []string{"sub-1"}).Return([]*rejudge.Rejudge{}, nil).Once()
		svc.On("RequestRejudgeSubmissions", int32(1), submissions, "Test reason").Return().Once()

		output, err := uc.Execute(validInput())

		assert.NoError(t, err)
		assert.NotNil(t, output)
	})

	t.Run("should rejudge with specific grading items", func(t *testing.T) {
		repo := mocks.NewSubmissionRepository(t)
		rejudgeRepo := mocks.NewRejudgeRepository(t)
		svc := mocks.NewSubmissionService(t)
		uc := rejudgecontestproject.NewUsecase(repo, rejudgeRepo, svc)

		gradingItemIds := []int32{1, 2, 3}
		submissions := []*submission.Submission{
			submission.New(&submission.SubmissionConfig{ID: "sub-1", ContestProjectID: 1}),
		}
		repo.On("FindByContestProjectIDAndGradingItemIDs", int32(1), gradingItemIds).Return(submissions, nil).Once()
		rejudgeRepo.On("FindAwaitingDecisionBySubmissionIDs", []string{"sub-1"}).Return([]*rejudge.Rejudge{}, nil).Once()
		svc.On("RequestRejudgeSubmissions", int32(1), submissions, "Test reason").Return().Once()

		input := validInput()
		input.GradingItemIds = &gradingItemIds

		output, err := uc.Execute(input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
	})

	// The contest project has to travel with the request rather than be read off
	// the submissions downstream: a project with nothing submitted to it yet is a
	// legitimate request, and the rejudge it produces still has to record which
	// project it belongs to.
	t.Run("passes the contest project on even when the project has no submissions", func(t *testing.T) {
		repo := mocks.NewSubmissionRepository(t)
		rejudgeRepo := mocks.NewRejudgeRepository(t)
		svc := mocks.NewSubmissionService(t)
		uc := rejudgecontestproject.NewUsecase(repo, rejudgeRepo, svc)

		none := []*submission.Submission{}
		repo.On("FindByContestProjectIDAndGradingItemIDs", int32(1), []int32{}).Return(none, nil).Once()
		rejudgeRepo.On("FindAwaitingDecisionBySubmissionIDs", []string{}).Return([]*rejudge.Rejudge{}, nil).Once()
		svc.On("RequestRejudgeSubmissions", int32(1), none, "Test reason").Return().Once()

		output, err := uc.Execute(validInput())

		assert.NoError(t, err)
		assert.NotNil(t, output)
	})

	t.Run("should return error when contest project ID is invalid", func(t *testing.T) {
		repo := mocks.NewSubmissionRepository(t)
		rejudgeRepo := mocks.NewRejudgeRepository(t)
		svc := mocks.NewSubmissionService(t)
		uc := rejudgecontestproject.NewUsecase(repo, rejudgeRepo, svc)

		input := validInput()
		input.ContestProjectID = 0

		_, err := uc.Execute(input)

		assert.ErrorIs(t, err, rejudgecontestproject.ErrorInvalidContestProjectID)
	})

	t.Run("should return error when reason is empty", func(t *testing.T) {
		repo := mocks.NewSubmissionRepository(t)
		rejudgeRepo := mocks.NewRejudgeRepository(t)
		svc := mocks.NewSubmissionService(t)
		uc := rejudgecontestproject.NewUsecase(repo, rejudgeRepo, svc)

		input := validInput()
		input.Reason = ""

		_, err := uc.Execute(input)

		assert.ErrorIs(t, err, rejudgecontestproject.ErrorRequiredReason)
	})

	t.Run("should return error when FindByContestProjectIDAndGradingItemIDs fails", func(t *testing.T) {
		repo := mocks.NewSubmissionRepository(t)
		rejudgeRepo := mocks.NewRejudgeRepository(t)
		svc := mocks.NewSubmissionService(t)
		uc := rejudgecontestproject.NewUsecase(repo, rejudgeRepo, svc)

		repo.On("FindByContestProjectIDAndGradingItemIDs", int32(1), testifymock.Anything).Return(([]*submission.Submission)(nil), assert.AnError).Once()

		_, err := uc.Execute(validInput())

		assert.Error(t, err)
	})

	t.Run("refuses the request when any submission has a rejudge awaiting a decision", func(t *testing.T) {
		repo := mocks.NewSubmissionRepository(t)
		rejudgeRepo := mocks.NewRejudgeRepository(t)
		svc := mocks.NewSubmissionService(t)
		uc := rejudgecontestproject.NewUsecase(repo, rejudgeRepo, svc)

		submissions := []*submission.Submission{
			submission.New(&submission.SubmissionConfig{ID: "sub-1", ContestProjectID: 1}),
			submission.New(&submission.SubmissionConfig{ID: "sub-2", ContestProjectID: 1}),
		}
		outstanding, err := rejudge.New(&rejudge.RejudgeConfig{ID: "rejudge-1", Reason: "earlier", Status: "Running"})
		assert.NoError(t, err)

		repo.On("FindByContestProjectIDAndGradingItemIDs", int32(1), []int32{}).Return(submissions, nil).Once()
		rejudgeRepo.On("FindAwaitingDecisionBySubmissionIDs", []string{"sub-1", "sub-2"}).
			Return([]*rejudge.Rejudge{outstanding}, nil).Once()

		_, err = uc.Execute(validInput())

		assert.ErrorIs(t, err, rejudgecontestproject.ErrorRejudgeAlreadyAwaitingDecision)
		svc.AssertNotCalled(t, "RequestRejudgeSubmissions")
	})
}
