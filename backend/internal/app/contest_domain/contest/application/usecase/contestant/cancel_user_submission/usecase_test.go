package cancelusersubmission_test

import (
	cancelusersubmission "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/cancel_user_submission"
	"judge-backend/internal/app/contest_domain/contest/domain/entity/submission"
	"judge-backend/internal/app/contest_domain/contest/mocks"
	sharedtypes "judge-backend/internal/app/shared/types"
	"testing"

	"github.com/stretchr/testify/assert"
	testifymock "github.com/stretchr/testify/mock"
)

type noopLogger struct{}

func (l *noopLogger) Debugf(msg string, args ...interface{}) {}
func (l *noopLogger) Infof(msg string, args ...interface{})  {}
func (l *noopLogger) Warnf(msg string, args ...interface{})  {}
func (l *noopLogger) Errorf(msg string, args ...interface{}) {}
func (l *noopLogger) Fatalf(msg string, args ...interface{}) {}
func (l *noopLogger) Panicf(msg string, args ...interface{}) {}

var _ sharedtypes.Logger = (*noopLogger)(nil)

type UsecaseMocks struct {
	repository *mocks.SubmissionRepository
	service    *mocks.SubmissionService
}

func NewUsecaseMocks(t *testing.T) *UsecaseMocks {
	return &UsecaseMocks{
		repository: mocks.NewSubmissionRepository(t),
		service:    mocks.NewSubmissionService(t),
	}
}

func NewUsecase(t *testing.T) (cancelusersubmission.Usecase, *UsecaseMocks) {
	m := NewUsecaseMocks(t)
	return cancelusersubmission.NewUsecase(m.repository, m.service, &noopLogger{}), m
}

func newCancellableSubmission(t *testing.T, userID int32) *submission.Submission {
	t.Helper()
	s, err := submission.NewSubmission(&submission.SubmissionConfig{
		ID:         "sub-001",
		UserID:     userID,
		JudgmentID: "judge-001",
		Status:     "Queued",
	})
	assert.NoError(t, err)
	return s
}

func newNonCancellableSubmission(t *testing.T, userID int32) *submission.Submission {
	t.Helper()
	s, err := submission.NewSubmission(&submission.SubmissionConfig{
		ID:         "sub-001",
		UserID:     userID,
		JudgmentID: "judge-001",
		Status:     "Judged",
	})
	assert.NoError(t, err)
	return s
}

func TestUsecase_Execute(t *testing.T) {
	t.Run("should cancel submission successfully", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		sub := newCancellableSubmission(t, 1)
		m.repository.On("FindByID", "sub-001").Return(sub, nil).Once()
		m.service.On("RequestCancelSubmission", testifymock.Anything).Return().Once()

		_, err := usecase.Execute(&cancelusersubmission.Input{
			UserID:       1,
			SubmissionID: "sub-001",
		})

		assert.NoError(t, err)
	})

	t.Run("should return error when user id is invalid", func(t *testing.T) {
		usecase, _ := NewUsecase(t)

		_, err := usecase.Execute(&cancelusersubmission.Input{
			UserID:       0,
			SubmissionID: "sub-001",
		})

		assert.ErrorIs(t, err, cancelusersubmission.ErrorInvalidUserID)
	})

	t.Run("should return error when submission id is empty", func(t *testing.T) {
		usecase, _ := NewUsecase(t)

		_, err := usecase.Execute(&cancelusersubmission.Input{
			UserID:       1,
			SubmissionID: "",
		})

		assert.ErrorIs(t, err, cancelusersubmission.ErrorInvalidSubmissionID)
	})

	t.Run("should return error when submission not found", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		m.repository.On("FindByID", "sub-001").Return(nil, assert.AnError).Once()

		_, err := usecase.Execute(&cancelusersubmission.Input{
			UserID:       1,
			SubmissionID: "sub-001",
		})

		assert.Error(t, err)
	})

	t.Run("should return forbidden when user does not own submission", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		sub := newCancellableSubmission(t, 2)
		m.repository.On("FindByID", "sub-001").Return(sub, nil).Once()

		_, err := usecase.Execute(&cancelusersubmission.Input{
			UserID:       1,
			SubmissionID: "sub-001",
		})

		assert.ErrorIs(t, err, cancelusersubmission.ErrorForbidden)
	})

	t.Run("should return error when submission cannot be cancelled", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		sub := newNonCancellableSubmission(t, 1)
		m.repository.On("FindByID", "sub-001").Return(sub, nil).Once()

		_, err := usecase.Execute(&cancelusersubmission.Input{
			UserID:       1,
			SubmissionID: "sub-001",
		})

		assert.ErrorIs(t, err, cancelusersubmission.ErrorInvalidSubmissionStatus)
	})
}
