package replyclarification_test

import (
	replyclarification "judge-backend/internal/app/contest_domain/activity/application/usecase/clarification/reply_clarification"
	"judge-backend/internal/app/contest_domain/activity/domain/entity/clarification"
	"judge-backend/internal/app/contest_domain/activity/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	testifymock "github.com/stretchr/testify/mock"
)

func newClarification(t *testing.T) *clarification.Clarification {
	t.Helper()
	c, err := clarification.New(&clarification.ClarificationConfig{
		ContestProjectID: 1,
		Subject:          "Test subject",
		Content:          "Test content",
	})
	if err != nil {
		t.Fatalf("failed to create clarification: %v", err)
	}
	return c
}

func TestUsecase_Execute(t *testing.T) {
	t.Run("should reply clarification successfully", func(t *testing.T) {
		repo := mocks.NewClarificationRepository(t)
		svc := mocks.NewClarificationService(t)
		uc := replyclarification.NewUsecase(repo, svc)

		c := newClarification(t)
		repo.On("FindByID", int32(1)).Return(c, nil).Once()
		repo.On("Save", testifymock.Anything).Return(nil).Once()
		svc.On("NotifyClarificationToUser", testifymock.Anything).Return().Once()

		output, err := uc.Execute(&replyclarification.Input{
			ID:      1,
			Content: "Reply content",
		})

		assert.NoError(t, err)
		assert.NotNil(t, output)
	})

	t.Run("should return error when clarification ID is invalid", func(t *testing.T) {
		repo := mocks.NewClarificationRepository(t)
		svc := mocks.NewClarificationService(t)
		uc := replyclarification.NewUsecase(repo, svc)

		_, err := uc.Execute(&replyclarification.Input{
			ID:      0,
			Content: "Reply content",
		})

		assert.ErrorIs(t, err, replyclarification.ErrorInvalidClarificationID)
	})

	t.Run("should return error when content is empty", func(t *testing.T) {
		repo := mocks.NewClarificationRepository(t)
		svc := mocks.NewClarificationService(t)
		uc := replyclarification.NewUsecase(repo, svc)

		_, err := uc.Execute(&replyclarification.Input{
			ID:      1,
			Content: "",
		})

		assert.ErrorIs(t, err, replyclarification.ErrorContentIsRequired)
	})

	t.Run("should return error when FindByID fails", func(t *testing.T) {
		repo := mocks.NewClarificationRepository(t)
		svc := mocks.NewClarificationService(t)
		uc := replyclarification.NewUsecase(repo, svc)

		repo.On("FindByID", int32(1)).Return((*clarification.Clarification)(nil), assert.AnError).Once()

		_, err := uc.Execute(&replyclarification.Input{
			ID:      1,
			Content: "Reply content",
		})

		assert.Error(t, err)
	})

	t.Run("should return error when clarification not found", func(t *testing.T) {
		repo := mocks.NewClarificationRepository(t)
		svc := mocks.NewClarificationService(t)
		uc := replyclarification.NewUsecase(repo, svc)

		repo.On("FindByID", int32(1)).Return((*clarification.Clarification)(nil), nil).Once()

		_, err := uc.Execute(&replyclarification.Input{
			ID:      1,
			Content: "Reply content",
		})

		assert.ErrorIs(t, err, replyclarification.ErrorClarificationNotFound)
	})

	t.Run("should return error when Save fails", func(t *testing.T) {
		repo := mocks.NewClarificationRepository(t)
		svc := mocks.NewClarificationService(t)
		uc := replyclarification.NewUsecase(repo, svc)

		c := newClarification(t)
		repo.On("FindByID", int32(1)).Return(c, nil).Once()
		repo.On("Save", testifymock.Anything).Return(assert.AnError).Once()

		_, err := uc.Execute(&replyclarification.Input{
			ID:      1,
			Content: "Reply content",
		})

		assert.Error(t, err)
	})
}
