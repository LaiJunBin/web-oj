package unignoresubmission_test

import (
	unignoresubmission "judge-backend/internal/app/contest_domain/activity/application/usecase/submission/unignore_submission"
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
		Ignored:          true,
	})
}

func TestUsecase_Execute(t *testing.T) {
	t.Run("should unignore submission successfully", func(t *testing.T) {
		repo := mocks.NewSubmissionRepository(t)
		uc := unignoresubmission.NewUsecase(repo)

		sub := newSubmission()
		repo.On("FindByID", "sub-1").Return(sub, nil).Once()
		repo.On("Save", testifymock.Anything).Return(nil).Once()

		output, err := uc.Execute(&unignoresubmission.Input{SubmissionID: "sub-1"})

		assert.NoError(t, err)
		assert.NotNil(t, output)
	})

	t.Run("should return error when submission ID is empty", func(t *testing.T) {
		repo := mocks.NewSubmissionRepository(t)
		uc := unignoresubmission.NewUsecase(repo)

		_, err := uc.Execute(&unignoresubmission.Input{SubmissionID: ""})

		assert.ErrorIs(t, err, unignoresubmission.ErrorRequiredSubmissionID)
	})

	t.Run("should return error when FindByID fails", func(t *testing.T) {
		repo := mocks.NewSubmissionRepository(t)
		uc := unignoresubmission.NewUsecase(repo)

		repo.On("FindByID", "sub-1").Return((*submission.Submission)(nil), assert.AnError).Once()

		_, err := uc.Execute(&unignoresubmission.Input{SubmissionID: "sub-1"})

		assert.Error(t, err)
	})

	t.Run("should return error when submission not found", func(t *testing.T) {
		repo := mocks.NewSubmissionRepository(t)
		uc := unignoresubmission.NewUsecase(repo)

		repo.On("FindByID", "sub-1").Return((*submission.Submission)(nil), nil).Once()

		_, err := uc.Execute(&unignoresubmission.Input{SubmissionID: "sub-1"})

		assert.ErrorIs(t, err, unignoresubmission.ErrorSubmissionNotFound)
	})

	t.Run("should return internal server error when Save fails", func(t *testing.T) {
		repo := mocks.NewSubmissionRepository(t)
		uc := unignoresubmission.NewUsecase(repo)

		sub := newSubmission()
		repo.On("FindByID", "sub-1").Return(sub, nil).Once()
		repo.On("Save", testifymock.Anything).Return(assert.AnError).Once()

		_, err := uc.Execute(&unignoresubmission.Input{SubmissionID: "sub-1"})

		assert.ErrorIs(t, err, unignoresubmission.ErrorInternalServer)
	})
}
