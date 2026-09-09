package listuserprojectsubmissions_test

import (
	"testing"

	"judge-backend/internal/app/contest_domain/contest/application/query/dto"
	listuserprojectsubmissions "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/list_user_project_submissions"
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

type UsecaseMocks struct {
	query *mocks.SubmissionQuery
}

func NewUsecaseMocks(t *testing.T) *UsecaseMocks {
	return &UsecaseMocks{
		query: mocks.NewSubmissionQuery(t),
	}
}

func NewUsecase(t *testing.T) (listuserprojectsubmissions.Usecase, *UsecaseMocks) {
	m := NewUsecaseMocks(t)
	return listuserprojectsubmissions.NewUsecase(m.query, &noopLogger{}), m
}

func TestUsecase_Execute(t *testing.T) {
	t.Run("should return submissions for user project", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		submissions := []*dto.Submission{
			{ID: "sub-001", Status: "Accepted", Score: 100},
			{ID: "sub-002", Status: "WrongAnswer", Score: 0},
		}
		m.query.On("ListUserProjectSubmissionsByProjectID", int32(1), int32(10)).Return(submissions, nil).Once()

		output, err := usecase.Execute(&listuserprojectsubmissions.Input{
			UserID:           1,
			ContestProjectID: 10,
		})

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, submissions, output.Submissions)
		assert.Len(t, output.Submissions, 2)
	})

	t.Run("should return empty list when no submissions exist", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		m.query.On("ListUserProjectSubmissionsByProjectID", int32(1), int32(10)).Return([]*dto.Submission{}, nil).Once()

		output, err := usecase.Execute(&listuserprojectsubmissions.Input{
			UserID:           1,
			ContestProjectID: 10,
		})

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Empty(t, output.Submissions)
	})

	t.Run("should return error when query fails", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		m.query.On("ListUserProjectSubmissionsByProjectID", int32(1), int32(10)).Return(([]*dto.Submission)(nil), assert.AnError).Once()

		output, err := usecase.Execute(&listuserprojectsubmissions.Input{
			UserID:           1,
			ContestProjectID: 10,
		})

		assert.Error(t, err)
		assert.Nil(t, output)
	})
}
