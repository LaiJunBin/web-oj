package getusersubmission_test

import (
	"errors"
	"testing"

	"judge-backend/internal/app/contest_domain/contest/application/query/dto"
	getusersubmission "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/get_user_submission"
	"judge-backend/internal/app/contest_domain/contest/mocks"
	sharedtypes "judge-backend/internal/app/shared/types"

	"github.com/stretchr/testify/assert"
)

type noopLogger struct{}

func (l *noopLogger) Debugf(msg string, args ...interface{}) {}
func (l *noopLogger) Infof(msg string, args ...interface{})  {}
func (l *noopLogger) Warnf(msg string, args ...interface{})  {}
func (l *noopLogger) Errorf(msg string, args ...interface{}) {}
func (l *noopLogger) Fatalf(msg string, args ...interface{}) {}
func (l *noopLogger) Panicf(msg string, args ...interface{}) {}

var _ sharedtypes.Logger = (*noopLogger)(nil)

type usecaseMocks struct {
	query *mocks.SubmissionQuery
}

func newUsecaseMocks(t *testing.T) *usecaseMocks {
	return &usecaseMocks{
		query: mocks.NewSubmissionQuery(t),
	}
}

func newUsecase(t *testing.T) (getusersubmission.Usecase, *usecaseMocks) {
	m := newUsecaseMocks(t)
	return getusersubmission.NewUsecase(m.query, &noopLogger{}), m
}

func TestUsecase_Execute(t *testing.T) {
	t.Run("should return submission detail on success", func(t *testing.T) {
		usecase, m := newUsecase(t)

		expected := &dto.SubmissionDetail{
			Submission: dto.Submission{ID: "sub-abc-123", Status: "accepted", Score: 100.0},
		}
		m.query.On("GetUserSubmissionByID", int32(10), "sub-abc-123").Return(expected, nil).Once()

		output, err := usecase.Execute(&getusersubmission.Input{UserID: 10, SubmissionID: "sub-abc-123"})

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, expected, output.Submission)
	})

	t.Run("should return nil submission when query returns nil", func(t *testing.T) {
		usecase, m := newUsecase(t)

		m.query.On("GetUserSubmissionByID", int32(10), "sub-not-found").Return((*dto.SubmissionDetail)(nil), nil).Once()

		output, err := usecase.Execute(&getusersubmission.Input{UserID: 10, SubmissionID: "sub-not-found"})

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Nil(t, output.Submission)
	})

	t.Run("should return error when query fails", func(t *testing.T) {
		usecase, m := newUsecase(t)

		queryErr := errors.New("database error")
		m.query.On("GetUserSubmissionByID", int32(10), "sub-abc-123").Return((*dto.SubmissionDetail)(nil), queryErr).Once()

		output, err := usecase.Execute(&getusersubmission.Input{UserID: 10, SubmissionID: "sub-abc-123"})

		assert.Error(t, err)
		assert.ErrorIs(t, err, queryErr)
		assert.Nil(t, output)
	})
}
