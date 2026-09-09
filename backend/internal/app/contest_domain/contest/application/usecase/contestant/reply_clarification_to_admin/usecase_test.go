package replyclarificationtoadmin_test

import (
	"testing"

	replyclarificationtoadmin "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/reply_clarification_to_admin"
	"judge-backend/internal/app/contest_domain/contest/domain/entity/clarification"
	"judge-backend/internal/app/contest_domain/contest/mocks"
	sharedtypes "judge-backend/internal/app/shared/types"

	"github.com/stretchr/testify/assert"
	testifymock "github.com/stretchr/testify/mock"
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
	repository *mocks.ClarificationRepository
}

func NewUsecaseMocks(t *testing.T) *UsecaseMocks {
	return &UsecaseMocks{
		repository: mocks.NewClarificationRepository(t),
	}
}

func NewUsecase(t *testing.T) (replyclarificationtoadmin.Usecase, *UsecaseMocks) {
	m := NewUsecaseMocks(t)
	return replyclarificationtoadmin.NewUsecase(m.repository, &noopLogger{}), m
}

func newClarificationOwnedByUser(t *testing.T, senderUserID int32) *clarification.Clarification {
	t.Helper()
	c, err := clarification.New(&clarification.ClarificationConfig{
		ContestProjectID: 1,
		SenderUserID:     senderUserID,
		ReceiverUserID:   0,
		IsAdminBroadcast: false,
		Subject:          "Test Subject",
		Content:          "Test Content",
	})
	assert.NoError(t, err)
	return c
}

func newClarificationNotOwnedByUser(t *testing.T) *clarification.Clarification {
	t.Helper()
	// SenderUserID=2, ReceiverUserID=3, IsAdminBroadcast=false
	// so user with ID 1 cannot reply
	c, err := clarification.New(&clarification.ClarificationConfig{
		ContestProjectID: 1,
		SenderUserID:     2,
		ReceiverUserID:   3,
		IsAdminBroadcast: false,
		Subject:          "Other User Subject",
		Content:          "Other User Content",
	})
	assert.NoError(t, err)
	return c
}

func TestUsecase_Execute(t *testing.T) {
	t.Run("should reply clarification successfully when user is sender", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		c := newClarificationOwnedByUser(t, 1)
		m.repository.On("FindByID", int32(1)).Return(c, nil).Once()
		m.repository.On("Save", testifymock.AnythingOfType("*clarification.Clarification")).Return(nil).Once()

		output, err := usecase.Execute(&replyclarificationtoadmin.Input{
			UserID:          1,
			ClarificationID: 1,
			Content:         "Reply content",
		})

		assert.NoError(t, err)
		assert.NotNil(t, output)
	})

	t.Run("should return error when content is empty", func(t *testing.T) {
		usecase, _ := NewUsecase(t)

		output, err := usecase.Execute(&replyclarificationtoadmin.Input{
			UserID:          1,
			ClarificationID: 1,
			Content:         "",
		})

		assert.ErrorIs(t, err, replyclarificationtoadmin.ErrorContentIsRequired)
		assert.Nil(t, output)
	})

	t.Run("should return error when clarification not found", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		m.repository.On("FindByID", int32(99)).Return((*clarification.Clarification)(nil), nil).Once()

		output, err := usecase.Execute(&replyclarificationtoadmin.Input{
			UserID:          1,
			ClarificationID: 99,
			Content:         "Reply content",
		})

		assert.ErrorIs(t, err, replyclarificationtoadmin.ErrorClarificationNotFound)
		assert.Nil(t, output)
	})

	t.Run("should return error when FindByID fails", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		m.repository.On("FindByID", int32(1)).Return((*clarification.Clarification)(nil), assert.AnError).Once()

		output, err := usecase.Execute(&replyclarificationtoadmin.Input{
			UserID:          1,
			ClarificationID: 1,
			Content:         "Reply content",
		})

		assert.Error(t, err)
		assert.Nil(t, output)
	})

	t.Run("should return forbidden when user cannot reply", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		c := newClarificationNotOwnedByUser(t)
		m.repository.On("FindByID", int32(1)).Return(c, nil).Once()

		output, err := usecase.Execute(&replyclarificationtoadmin.Input{
			UserID:          1,
			ClarificationID: 1,
			Content:         "Reply content",
		})

		assert.ErrorIs(t, err, replyclarificationtoadmin.ErrorClarificationForbidden)
		assert.Nil(t, output)
	})

	t.Run("should return error when Save fails", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		c := newClarificationOwnedByUser(t, 1)
		m.repository.On("FindByID", int32(1)).Return(c, nil).Once()
		m.repository.On("Save", testifymock.AnythingOfType("*clarification.Clarification")).Return(assert.AnError).Once()

		output, err := usecase.Execute(&replyclarificationtoadmin.Input{
			UserID:          1,
			ClarificationID: 1,
			Content:         "Reply content",
		})

		assert.Error(t, err)
		assert.Nil(t, output)
	})
}
