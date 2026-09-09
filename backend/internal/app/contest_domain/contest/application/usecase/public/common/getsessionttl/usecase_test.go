package getsessionttl_test

import (
	"testing"

	getsessionttl "judge-backend/internal/app/contest_domain/contest/application/usecase/public/common/getsessionttl"
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
	query *mocks.SessionQuery
}

func newUsecaseMocks(t *testing.T) *usecaseMocks {
	return &usecaseMocks{
		query: mocks.NewSessionQuery(t),
	}
}

func newUsecase(t *testing.T) (getsessionttl.Usecase, *usecaseMocks) {
	m := newUsecaseMocks(t)
	return getsessionttl.NewUsecase(m.query, &noopLogger{}), m
}

func TestUsecase_Execute(t *testing.T) {
	t.Run("should return TTL successfully", func(t *testing.T) {
		usecase, m := newUsecase(t)

		m.query.On("GetTTL", "session-123").Return(300).Once()

		output, err := usecase.Execute(&getsessionttl.Input{SessionID: "session-123"})

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, 300, output.TTL)
	})

	t.Run("should return error when session ID is empty", func(t *testing.T) {
		usecase, _ := newUsecase(t)

		output, err := usecase.Execute(&getsessionttl.Input{SessionID: ""})

		assert.Error(t, err)
		assert.ErrorIs(t, err, getsessionttl.ErrorInvalidSessionID)
		assert.Nil(t, output)
	})
}
