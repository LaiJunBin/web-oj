package sendclarification_test

import (
	sendclarification "judge-backend/internal/app/contest_domain/activity/application/usecase/clarification/send_clarification"
	"judge-backend/internal/app/contest_domain/activity/domain/repository"
	"judge-backend/internal/app/contest_domain/activity/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	testifymock "github.com/stretchr/testify/mock"
)

func ptr[T any](v T) *T {
	return &v
}

func validInput() *sendclarification.Input {
	return &sendclarification.Input{
		ContestProjectID: 1,
		SenderUserID:     10,
		ReceiverUserID:   ptr(int32(20)),
		Subject:          "Test subject",
		Content:          "Test content",
	}
}

func TestUsecase_Execute(t *testing.T) {
	t.Run("should send clarification successfully", func(t *testing.T) {
		repo := mocks.NewClarificationRepository(t)
		svc := mocks.NewClarificationService(t)
		uc := sendclarification.NewUsecase(repo, svc)

		repo.On("Save", testifymock.Anything).Return(nil).Once()
		svc.On("NotifyClarificationToUser", testifymock.Anything).Return().Once()

		output, err := uc.Execute(validInput())

		assert.NoError(t, err)
		assert.NotNil(t, output)
	})

	t.Run("should send clarification to all", func(t *testing.T) {
		repo := mocks.NewClarificationRepository(t)
		svc := mocks.NewClarificationService(t)
		uc := sendclarification.NewUsecase(repo, svc)

		repo.On("Save", testifymock.Anything).Return(nil).Once()
		svc.On("NotifyClarificationToUser", testifymock.Anything).Return().Once()

		input := &sendclarification.Input{
			ContestProjectID: 1,
			SenderUserID:     10,
			Subject:          "Broadcast subject",
			Content:          "Broadcast content",
			SendToAll:        true,
		}

		output, err := uc.Execute(input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
	})

	t.Run("should return error when contest project ID is invalid", func(t *testing.T) {
		repo := mocks.NewClarificationRepository(t)
		svc := mocks.NewClarificationService(t)
		uc := sendclarification.NewUsecase(repo, svc)

		input := validInput()
		input.ContestProjectID = 0

		_, err := uc.Execute(input)

		assert.ErrorIs(t, err, sendclarification.ErrorInvalidContestProjectID)
	})

	t.Run("should return error when subject is empty", func(t *testing.T) {
		repo := mocks.NewClarificationRepository(t)
		svc := mocks.NewClarificationService(t)
		uc := sendclarification.NewUsecase(repo, svc)

		input := validInput()
		input.Subject = ""

		_, err := uc.Execute(input)

		assert.ErrorIs(t, err, sendclarification.ErrorSubjectIsRequired)
	})

	t.Run("should return error when receiver and send to all are both missing", func(t *testing.T) {
		repo := mocks.NewClarificationRepository(t)
		svc := mocks.NewClarificationService(t)
		uc := sendclarification.NewUsecase(repo, svc)

		input := validInput()
		input.ReceiverUserID = nil
		input.SendToAll = false

		_, err := uc.Execute(input)

		assert.ErrorIs(t, err, sendclarification.ErrorReceiverUserIDOrSendToAllIsRequired)
	})

	t.Run("should return error when contest project not found", func(t *testing.T) {
		repo := mocks.NewClarificationRepository(t)
		svc := mocks.NewClarificationService(t)
		uc := sendclarification.NewUsecase(repo, svc)

		repo.On("Save", testifymock.Anything).Return(repository.ErrorContestProjectNotFound).Once()

		_, err := uc.Execute(validInput())

		assert.ErrorIs(t, err, sendclarification.ErrorContestProjectNotFound)
	})

	t.Run("should return error when receiver user not found", func(t *testing.T) {
		repo := mocks.NewClarificationRepository(t)
		svc := mocks.NewClarificationService(t)
		uc := sendclarification.NewUsecase(repo, svc)

		repo.On("Save", testifymock.Anything).Return(repository.ErrorReceiverUserNotFound).Once()

		_, err := uc.Execute(validInput())

		assert.ErrorIs(t, err, sendclarification.ErrorReceiverUserIDIsNotExists)
	})

	t.Run("should propagate unexpected save error", func(t *testing.T) {
		repo := mocks.NewClarificationRepository(t)
		svc := mocks.NewClarificationService(t)
		uc := sendclarification.NewUsecase(repo, svc)

		repo.On("Save", testifymock.Anything).Return(assert.AnError).Once()

		_, err := uc.Execute(validInput())

		assert.Error(t, err)
	})
}
