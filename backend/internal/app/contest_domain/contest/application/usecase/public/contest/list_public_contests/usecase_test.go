package listpubliccontests_test

import (
	"errors"
	"testing"

	"judge-backend/internal/app/contest_domain/contest/application/query/dto"
	listpubliccontests "judge-backend/internal/app/contest_domain/contest/application/usecase/public/contest/list_public_contests"
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
	query *mocks.ContestQuery
}

func newUsecaseMocks(t *testing.T) *usecaseMocks {
	return &usecaseMocks{
		query: mocks.NewContestQuery(t),
	}
}

func newUsecase(t *testing.T) (listpubliccontests.Usecase, *usecaseMocks) {
	m := newUsecaseMocks(t)
	return listpubliccontests.NewUsecase(m.query, &noopLogger{}), m
}

func TestUsecase_Execute(t *testing.T) {
	t.Run("should return list of public contests", func(t *testing.T) {
		usecase, m := newUsecase(t)

		expected := []dto.Contest{
			{ID: 1, Name: "Contest A"},
			{ID: 2, Name: "Contest B"},
		}
		m.query.On("ListPublicContests").Return(expected, nil).Once()

		output, err := usecase.Execute(&listpubliccontests.Input{})

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, expected, output.Contests)
	})

	t.Run("should return empty list when no public contests", func(t *testing.T) {
		usecase, m := newUsecase(t)

		m.query.On("ListPublicContests").Return([]dto.Contest{}, nil).Once()

		output, err := usecase.Execute(&listpubliccontests.Input{})

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Empty(t, output.Contests)
	})

	t.Run("should return error when query fails", func(t *testing.T) {
		usecase, m := newUsecase(t)

		queryErr := errors.New("database error")
		m.query.On("ListPublicContests").Return(nil, queryErr).Once()

		output, err := usecase.Execute(&listpubliccontests.Input{})

		assert.Error(t, err)
		assert.ErrorIs(t, err, queryErr)
		assert.Nil(t, output)
	})
}
