package updateclarification_test

import (
	updateclarification "judge-backend/internal/app/contest_domain/activity/application/usecase/clarification/update_clarification"
	"judge-backend/internal/app/contest_domain/activity/domain/entity/clarification"
	"judge-backend/internal/app/contest_domain/activity/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	testifymock "github.com/stretchr/testify/mock"
)

func ptr[T any](v T) *T {
	return &v
}

func newClarification(t *testing.T) *clarification.Clarification {
	t.Helper()
	c, err := clarification.New(&clarification.ClarificationConfig{
		ContestProjectID: 1,
		Subject:          "Test subject",
		Content:          "Test content",
		IsAnswered:       false,
	})
	if err != nil {
		t.Fatalf("failed to create clarification: %v", err)
	}
	return c
}

func TestUsecase_Execute(t *testing.T) {
	t.Run("should return error when IsAnswered is nil", func(t *testing.T) {
		repo := mocks.NewClarificationRepository(t)
		uc := updateclarification.NewUsecase(repo)

		_, err := uc.Execute(&updateclarification.Input{
			ID:         1,
			IsAnswered: nil,
		})

		assert.ErrorIs(t, err, updateclarification.ErrorIsAnsweredIsRequired)
	})

	t.Run("should return error when clarification not found", func(t *testing.T) {
		repo := mocks.NewClarificationRepository(t)
		uc := updateclarification.NewUsecase(repo)

		repo.On("FindByID", int32(1)).Return((*clarification.Clarification)(nil), nil).Once()

		_, err := uc.Execute(&updateclarification.Input{
			ID:         1,
			IsAnswered: ptr(true),
		})

		assert.ErrorIs(t, err, updateclarification.ErrorClarificationNotFound)
	})

	t.Run("should return error when FindByID fails", func(t *testing.T) {
		repo := mocks.NewClarificationRepository(t)
		uc := updateclarification.NewUsecase(repo)

		repo.On("FindByID", int32(1)).Return((*clarification.Clarification)(nil), assert.AnError).Once()

		_, err := uc.Execute(&updateclarification.Input{
			ID:         1,
			IsAnswered: ptr(true),
		})

		assert.Error(t, err)
	})

	t.Run("should return without saving when IsAnswered value is unchanged", func(t *testing.T) {
		repo := mocks.NewClarificationRepository(t)
		uc := updateclarification.NewUsecase(repo)

		c := newClarification(t) // IsAnswered = false
		repo.On("FindByID", int32(1)).Return(c, nil).Once()

		// no Save call expected
		output, err := uc.Execute(&updateclarification.Input{
			ID:         1,
			IsAnswered: ptr(false), // same value
		})

		assert.NoError(t, err)
		assert.NotNil(t, output)
	})

	t.Run("should update IsAnswered successfully", func(t *testing.T) {
		repo := mocks.NewClarificationRepository(t)
		uc := updateclarification.NewUsecase(repo)

		c := newClarification(t) // IsAnswered = false
		repo.On("FindByID", int32(1)).Return(c, nil).Once()
		repo.On("Save", testifymock.Anything).Return(nil).Once()

		output, err := uc.Execute(&updateclarification.Input{
			ID:         1,
			IsAnswered: ptr(true), // change to true
		})

		assert.NoError(t, err)
		assert.NotNil(t, output)
	})

	t.Run("should return error when Save fails", func(t *testing.T) {
		repo := mocks.NewClarificationRepository(t)
		uc := updateclarification.NewUsecase(repo)

		c := newClarification(t)
		repo.On("FindByID", int32(1)).Return(c, nil).Once()
		repo.On("Save", testifymock.Anything).Return(assert.AnError).Once()

		_, err := uc.Execute(&updateclarification.Input{
			ID:         1,
			IsAnswered: ptr(true),
		})

		assert.Error(t, err)
	})
}
