package ignoresubmission_test

import (
	ignoresubmission "judge-backend/internal/app/contest_domain/activity/application/usecase/submission/ignore_submission"
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
	t.Run("should ignore submission successfully", func(t *testing.T) {
		repo := mocks.NewSubmissionRepository(t)
		uc := ignoresubmission.NewUsecase(repo)

		sub := newSubmission()
		repo.On("FindByID", "sub-1").Return(sub, nil).Once()
		repo.On("Save", testifymock.Anything).Return(nil).Once()

		output, err := uc.Execute(&ignoresubmission.Input{SubmissionID: "sub-1"})

		assert.NoError(t, err)
		assert.NotNil(t, output)
	})

	t.Run("should return error when submission ID is empty", func(t *testing.T) {
		repo := mocks.NewSubmissionRepository(t)
		uc := ignoresubmission.NewUsecase(repo)

		_, err := uc.Execute(&ignoresubmission.Input{SubmissionID: ""})

		assert.ErrorIs(t, err, ignoresubmission.ErrorRequiredSubmissionID)
	})

	t.Run("should return error when FindByID fails", func(t *testing.T) {
		repo := mocks.NewSubmissionRepository(t)
		uc := ignoresubmission.NewUsecase(repo)

		repo.On("FindByID", "sub-1").Return((*submission.Submission)(nil), assert.AnError).Once()

		_, err := uc.Execute(&ignoresubmission.Input{SubmissionID: "sub-1"})

		assert.Error(t, err)
	})

	t.Run("should return error when submission not found", func(t *testing.T) {
		repo := mocks.NewSubmissionRepository(t)
		uc := ignoresubmission.NewUsecase(repo)

		repo.On("FindByID", "sub-1").Return((*submission.Submission)(nil), nil).Once()

		_, err := uc.Execute(&ignoresubmission.Input{SubmissionID: "sub-1"})

		assert.ErrorIs(t, err, ignoresubmission.ErrorSubmissionNotFound)
	})

	t.Run("should return internal server error when Save fails", func(t *testing.T) {
		repo := mocks.NewSubmissionRepository(t)
		uc := ignoresubmission.NewUsecase(repo)

		sub := newSubmission()
		repo.On("FindByID", "sub-1").Return(sub, nil).Once()
		repo.On("Save", testifymock.Anything).Return(assert.AnError).Once()

		_, err := uc.Execute(&ignoresubmission.Input{SubmissionID: "sub-1"})

		assert.ErrorIs(t, err, ignoresubmission.ErrorInternalServer)
	})
}
