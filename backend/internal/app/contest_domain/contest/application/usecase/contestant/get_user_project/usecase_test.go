package getuserproject_test

import (
	"errors"
	"testing"

	"judge-backend/internal/app/contest_domain/contest/application/query/dto"
	getuserproject "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/get_user_project"
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
	query *mocks.ProjectQuery
}

func newUsecaseMocks(t *testing.T) *usecaseMocks {
	return &usecaseMocks{
		query: mocks.NewProjectQuery(t),
	}
}

func newUsecase(t *testing.T) (getuserproject.Usecase, *usecaseMocks) {
	m := newUsecaseMocks(t)
	return getuserproject.NewUsecase(m.query, &noopLogger{}), m
}

func TestUsecase_Execute(t *testing.T) {
	t.Run("should return project detail on success", func(t *testing.T) {
		usecase, m := newUsecase(t)

		expected := dto.UserProjectDetail{
			ProjectDetail: dto.ProjectDetail{
				Project:     dto.Project{ID: 5, Name: "Test Project", Status: "active"},
				Description: "A test project",
			},
			IsParticipant: true,
		}
		m.query.On("GetUserProjectByID", int32(10), int32(5)).Return(expected, nil).Once()

		output, err := usecase.Execute(&getuserproject.Input{UserID: 10, ContestProjectID: 5})

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, expected, output.Project)
	})

	t.Run("should return error when query fails", func(t *testing.T) {
		usecase, m := newUsecase(t)

		queryErr := errors.New("database error")
		m.query.On("GetUserProjectByID", int32(10), int32(5)).Return(dto.UserProjectDetail{}, queryErr).Once()

		output, err := usecase.Execute(&getuserproject.Input{UserID: 10, ContestProjectID: 5})

		assert.Error(t, err)
		assert.ErrorIs(t, err, queryErr)
		assert.Nil(t, output)
	})
}
