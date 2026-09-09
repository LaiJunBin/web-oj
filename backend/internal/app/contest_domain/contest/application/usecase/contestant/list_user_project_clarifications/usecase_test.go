package listuserprojectclarifications_test

import (
	"testing"

	"judge-backend/internal/app/contest_domain/contest/application/query/dto"
	listuserprojectclarifications "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/list_user_project_clarifications"
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
	query               *mocks.ClarificationQuery
	contestProjectQuery *mocks.ContestProjectQuery
}

func NewUsecaseMocks(t *testing.T) *UsecaseMocks {
	return &UsecaseMocks{
		query:               mocks.NewClarificationQuery(t),
		contestProjectQuery: mocks.NewContestProjectQuery(t),
	}
}

func NewUsecase(t *testing.T) (listuserprojectclarifications.Usecase, *UsecaseMocks) {
	m := NewUsecaseMocks(t)
	return listuserprojectclarifications.NewUsecase(m.query, m.contestProjectQuery, &noopLogger{}), m
}

func TestUsecase_Execute(t *testing.T) {
	t.Run("should return clarifications for user project", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		clarifications := []*dto.Clarification{
			{ID: 1, Subject: "Question 1", Content: "Content 1"},
			{ID: 2, Subject: "Question 2", Content: "Content 2"},
		}
		m.contestProjectQuery.On("CheckUserCanAccessProject", int32(10), int32(1)).Return(nil).Once()
		m.query.On("ListUserClarificationsByProjectID", int32(1), int32(10)).Return(clarifications, nil).Once()

		output, err := usecase.Execute(&listuserprojectclarifications.Input{
			UserID:           1,
			ContestProjectID: 10,
		})

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, clarifications, output.Clarifications)
		assert.Len(t, output.Clarifications, 2)
	})

	t.Run("should return empty list when no clarifications exist", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		m.contestProjectQuery.On("CheckUserCanAccessProject", int32(10), int32(1)).Return(nil).Once()
		m.query.On("ListUserClarificationsByProjectID", int32(1), int32(10)).Return([]*dto.Clarification{}, nil).Once()

		output, err := usecase.Execute(&listuserprojectclarifications.Input{
			UserID:           1,
			ContestProjectID: 10,
		})

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Empty(t, output.Clarifications)
	})

	t.Run("should return error when user cannot access project", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		m.contestProjectQuery.On("CheckUserCanAccessProject", int32(10), int32(1)).Return(assert.AnError).Once()

		output, err := usecase.Execute(&listuserprojectclarifications.Input{
			UserID:           1,
			ContestProjectID: 10,
		})

		assert.Error(t, err)
		assert.Nil(t, output)
	})

	t.Run("should return error when query fails", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		m.contestProjectQuery.On("CheckUserCanAccessProject", int32(10), int32(1)).Return(nil).Once()
		m.query.On("ListUserClarificationsByProjectID", int32(1), int32(10)).Return(([]*dto.Clarification)(nil), assert.AnError).Once()

		output, err := usecase.Execute(&listuserprojectclarifications.Input{
			UserID:           1,
			ContestProjectID: 10,
		})

		assert.Error(t, err)
		assert.Nil(t, output)
	})
}
