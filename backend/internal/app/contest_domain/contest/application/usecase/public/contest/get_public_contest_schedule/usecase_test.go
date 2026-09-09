package getpubliccontestschedule_test

import (
	"errors"
	"testing"

	"judge-backend/internal/app/contest_domain/contest/application/query/dto"
	getpubliccontestschedule "judge-backend/internal/app/contest_domain/contest/application/usecase/public/contest/get_public_contest_schedule"
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

func newUsecase(t *testing.T) (getpubliccontestschedule.Usecase, *usecaseMocks) {
	m := newUsecaseMocks(t)
	return getpubliccontestschedule.NewUsecase(m.query, &noopLogger{}), m
}

func TestUsecase_Execute(t *testing.T) {
	t.Run("should return public contest schedule", func(t *testing.T) {
		usecase, m := newUsecase(t)

		expected := dto.ContestSchedule{
			IsPublic: true,
			Projects: []*dto.ContestScheduleProject{
				{Name: "Project A"},
			},
		}
		m.query.On("GetPublicContestScheduleByID", int32(1)).Return(expected, nil).Once()

		output, err := usecase.Execute(&getpubliccontestschedule.Input{ContestID: 1})

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, expected, output.ContestSchedule)
	})

	t.Run("should return error when contest ID is invalid", func(t *testing.T) {
		usecase, _ := newUsecase(t)

		output, err := usecase.Execute(&getpubliccontestschedule.Input{ContestID: 0})

		assert.Error(t, err)
		assert.ErrorIs(t, err, getpubliccontestschedule.ErrorInvalidContestID)
		assert.Nil(t, output)
	})

	t.Run("should return error when query fails", func(t *testing.T) {
		usecase, m := newUsecase(t)

		queryErr := errors.New("database error")
		m.query.On("GetPublicContestScheduleByID", int32(1)).Return(dto.ContestSchedule{}, queryErr).Once()

		output, err := usecase.Execute(&getpubliccontestschedule.Input{ContestID: 1})

		assert.Error(t, err)
		assert.ErrorIs(t, err, queryErr)
		assert.Nil(t, output)
	})

	t.Run("should return error when contest is not public", func(t *testing.T) {
		usecase, m := newUsecase(t)

		schedule := dto.ContestSchedule{
			IsPublic: false,
			Projects: []*dto.ContestScheduleProject{},
		}
		m.query.On("GetPublicContestScheduleByID", int32(1)).Return(schedule, nil).Once()

		output, err := usecase.Execute(&getpubliccontestschedule.Input{ContestID: 1})

		assert.Error(t, err)
		assert.ErrorIs(t, err, getpubliccontestschedule.ErrorNotFoundContest)
		assert.Nil(t, output)
	})
}
