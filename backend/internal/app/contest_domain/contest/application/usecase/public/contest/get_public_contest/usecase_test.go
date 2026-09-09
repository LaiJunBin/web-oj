package getpubliccontest_test

import (
	"errors"
	"testing"

	"judge-backend/internal/app/contest_domain/contest/application/query/dto"
	getpubliccontest "judge-backend/internal/app/contest_domain/contest/application/usecase/public/contest/get_public_contest"
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

func newUsecase(t *testing.T) (getpubliccontest.Usecase, *usecaseMocks) {
	m := newUsecaseMocks(t)
	return getpubliccontest.NewUsecase(m.query, &noopLogger{}), m
}

func TestUsecase_Execute(t *testing.T) {
	t.Run("should return public contest successfully", func(t *testing.T) {
		usecase, m := newUsecase(t)

		expected := dto.ContestDetail{
			Contest: dto.Contest{ID: 1, Name: "Public Contest"},
		}
		m.query.On("GetPublicContestByID", int32(1)).Return(expected, nil).Once()

		output, err := usecase.Execute(&getpubliccontest.Input{ContestID: 1})

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, expected, output.Contest)
	})

	t.Run("should return error when query fails", func(t *testing.T) {
		usecase, m := newUsecase(t)

		queryErr := errors.New("database error")
		m.query.On("GetPublicContestByID", int32(1)).Return(dto.ContestDetail{}, queryErr).Once()

		output, err := usecase.Execute(&getpubliccontest.Input{ContestID: 1})

		assert.Error(t, err)
		assert.ErrorIs(t, err, queryErr)
		assert.Nil(t, output)
	})
}
