package getusercontestschedule_test

import (
	"errors"
	"testing"

	"judge-backend/internal/app/contest_domain/contest/application/query/dto"
	getusercontestschedule "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/get_user_contest_schedule"
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

func newUsecase(t *testing.T) (getusercontestschedule.Usecase, *usecaseMocks) {
	m := newUsecaseMocks(t)
	return getusercontestschedule.NewUsecase(m.query, &noopLogger{}), m
}

func TestUsecase_Execute(t *testing.T) {
	t.Run("should return contest schedule on success", func(t *testing.T) {
		usecase, m := newUsecase(t)

		expected := dto.ContestSchedule{
			IsPublic: true,
			Projects: nil,
		}
		m.query.On("GetUserContestScheduleByID", int32(7), int32(1)).Return(expected, nil).Once()

		output, err := usecase.Execute(&getusercontestschedule.Input{UserID: 7, ContestID: 1})

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, expected, output.ContestSchedule)
	})

	t.Run("should return error when contest id is zero", func(t *testing.T) {
		usecase, _ := newUsecase(t)

		output, err := usecase.Execute(&getusercontestschedule.Input{ContestID: 0})

		assert.Error(t, err)
		assert.ErrorIs(t, err, getusercontestschedule.ErrorInvalidContestID)
		assert.Nil(t, output)
	})

	t.Run("should return error when contest id is negative", func(t *testing.T) {
		usecase, _ := newUsecase(t)

		output, err := usecase.Execute(&getusercontestschedule.Input{ContestID: -5})

		assert.Error(t, err)
		assert.ErrorIs(t, err, getusercontestschedule.ErrorInvalidContestID)
		assert.Nil(t, output)
	})

	t.Run("should return error when query fails", func(t *testing.T) {
		usecase, m := newUsecase(t)

		queryErr := errors.New("database error")
		m.query.On("GetUserContestScheduleByID", int32(7), int32(1)).Return(dto.ContestSchedule{}, queryErr).Once()

		output, err := usecase.Execute(&getusercontestschedule.Input{UserID: 7, ContestID: 1})

		assert.Error(t, err)
		assert.ErrorIs(t, err, queryErr)
		assert.Nil(t, output)
	})
}
