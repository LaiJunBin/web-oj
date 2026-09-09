package sendclarificationtoadmin_test

import (
	"testing"

	sendclarificationtoadmin "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/send_clarification_to_admin"
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
	repository          *mocks.ClarificationRepository
	contestProjectQuery *mocks.ContestProjectQuery
}

func NewUsecaseMocks(t *testing.T) *UsecaseMocks {
	return &UsecaseMocks{
		repository:          mocks.NewClarificationRepository(t),
		contestProjectQuery: mocks.NewContestProjectQuery(t),
	}
}

func NewUsecase(t *testing.T) (sendclarificationtoadmin.Usecase, *UsecaseMocks) {
	m := NewUsecaseMocks(t)
	return sendclarificationtoadmin.NewUsecase(m.repository, m.contestProjectQuery, &noopLogger{}), m
}

func TestUsecase_Execute(t *testing.T) {
	t.Run("should send clarification successfully", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		m.contestProjectQuery.On("CheckUserCanAccessProject", int32(10), int32(1)).Return(nil).Once()
		m.repository.On("Save", testifymock.AnythingOfType("*clarification.Clarification")).Return(nil).Once()

		output, err := usecase.Execute(&sendclarificationtoadmin.Input{
			SenderUserID:     1,
			ContestProjectID: 10,
			Subject:          "Test Subject",
			Content:          "Test Content",
		})

		assert.NoError(t, err)
		assert.NotNil(t, output)
	})

	t.Run("should return error when subject is empty", func(t *testing.T) {
		usecase, _ := NewUsecase(t)

		output, err := usecase.Execute(&sendclarificationtoadmin.Input{
			SenderUserID:     1,
			ContestProjectID: 10,
			Subject:          "",
			Content:          "Test Content",
		})

		assert.ErrorIs(t, err, sendclarificationtoadmin.ErrorSubjectIsRequired)
		assert.Nil(t, output)
	})

	t.Run("should return error when content is empty", func(t *testing.T) {
		usecase, _ := NewUsecase(t)

		output, err := usecase.Execute(&sendclarificationtoadmin.Input{
			SenderUserID:     1,
			ContestProjectID: 10,
			Subject:          "Test Subject",
			Content:          "",
		})

		assert.ErrorIs(t, err, sendclarificationtoadmin.ErrorContentIsRequired)
		assert.Nil(t, output)
	})

	t.Run("should return error when user cannot access project", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		m.contestProjectQuery.On("CheckUserCanAccessProject", int32(10), int32(1)).Return(assert.AnError).Once()

		output, err := usecase.Execute(&sendclarificationtoadmin.Input{
			SenderUserID:     1,
			ContestProjectID: 10,
			Subject:          "Test Subject",
			Content:          "Test Content",
		})

		assert.Error(t, err)
		assert.Nil(t, output)
	})

	t.Run("should return error when repository save fails", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		m.contestProjectQuery.On("CheckUserCanAccessProject", int32(10), int32(1)).Return(nil).Once()
		m.repository.On("Save", testifymock.AnythingOfType("*clarification.Clarification")).Return(assert.AnError).Once()

		output, err := usecase.Execute(&sendclarificationtoadmin.Input{
			SenderUserID:     1,
			ContestProjectID: 10,
			Subject:          "Test Subject",
			Content:          "Test Content",
		})

		assert.Error(t, err)
		assert.Nil(t, output)
	})
}
