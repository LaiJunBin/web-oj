package getuserclarification_test

import (
	"judge-backend/internal/app/contest_domain/contest/application/query/dto"
	getuserclarification "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/get_user_clarification"
	"judge-backend/internal/app/contest_domain/contest/mocks"
	sharedtypes "judge-backend/internal/app/shared/types"
	"testing"

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
	query *mocks.ClarificationQuery
}

func NewUsecaseMocks(t *testing.T) *UsecaseMocks {
	return &UsecaseMocks{
		query: mocks.NewClarificationQuery(t),
	}
}

func NewUsecase(t *testing.T) (getuserclarification.Usecase, *UsecaseMocks) {
	m := NewUsecaseMocks(t)
	return getuserclarification.NewUsecase(m.query, &noopLogger{}), m
}

func TestUsecase_Execute(t *testing.T) {
	t.Run("should return clarification detail", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		detail := &dto.ClarificationDetail{Clarification: dto.Clarification{ID: 1, Subject: "Test"}}
		m.query.On("GetUserClarificationByID", int32(1), int32(1)).Return(detail, nil).Once()

		output, err := usecase.Execute(&getuserclarification.Input{UserID: 1, ClarificationID: 1})

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.NotNil(t, output.Clarification)
	})

	t.Run("should return error when clarification not found (nil result)", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		m.query.On("GetUserClarificationByID", int32(1), int32(99)).Return((*dto.ClarificationDetail)(nil), nil).Once()

		_, err := usecase.Execute(&getuserclarification.Input{UserID: 1, ClarificationID: 99})

		assert.ErrorIs(t, err, getuserclarification.ErrorClarificationNotFound)
	})

	t.Run("should return error when query fails", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		m.query.On("GetUserClarificationByID", int32(1), int32(1)).Return((*dto.ClarificationDetail)(nil), assert.AnError).Once()

		_, err := usecase.Execute(&getuserclarification.Input{UserID: 1, ClarificationID: 1})

		assert.Error(t, err)
	})
}
