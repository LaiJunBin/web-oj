package updatecontest_test

import (
	updatecontest "judge-backend/internal/app/contest_domain/contest/application/usecase/update_contest"
	"judge-backend/internal/app/contest_domain/contest/domain/entity/contest"
	"judge-backend/internal/app/contest_domain/contest/mocks"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	testifymock "github.com/stretchr/testify/mock"
)

type UsecaseMocks struct {
	repository *mocks.ContestRepository
	provider   *mocks.JudgePriorityStrategyValidator
}

func NewUsecaseMocks(t *testing.T) *UsecaseMocks {
	return &UsecaseMocks{
		repository: mocks.NewContestRepository(t),
		provider:   mocks.NewJudgePriorityStrategyValidator(t),
	}
}

func NewUsecase(t *testing.T) (updatecontest.Usecase, *UsecaseMocks) {
	m := NewUsecaseMocks(t)
	return updatecontest.NewUsecase(m.provider, m.repository), m
}

func newContest(t *testing.T) *contest.Contest {
	t.Helper()
	c, err := contest.New(&contest.ContestConfig{
		Name:             "Test Contest",
		Description:      "Description",
		ActivationTime:   time.Now(),
		DeactivationTime: time.Now().Add(24 * time.Hour),
		IsPublic:         true,
		Projects:         []contest.Project{},
	})
	assert.NoError(t, err)
	return c
}

func ptr[T any](v T) *T { return &v }

func TestUsecase_Execute(t *testing.T) {
	t.Run("should update contest name successfully", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		c := newContest(t)
		m.repository.On("FindByID", int32(1)).Return(c, nil).Once()
		m.repository.On("Save", testifymock.Anything).Return(nil).Once()

		_, err := usecase.Execute(&updatecontest.Input{
			ID:   1,
			Name: ptr("Updated Contest"),
		})

		assert.NoError(t, err)
	})

	t.Run("should return error when id is invalid", func(t *testing.T) {
		usecase, _ := NewUsecase(t)

		_, err := usecase.Execute(&updatecontest.Input{ID: 0, Name: ptr("New Name")})

		assert.ErrorIs(t, err, updatecontest.ErrorInvalidID)
	})

	t.Run("should return error when no fields to update", func(t *testing.T) {
		usecase, _ := NewUsecase(t)

		_, err := usecase.Execute(&updatecontest.Input{ID: 1})

		assert.ErrorIs(t, err, updatecontest.ErrorNoInput)
	})

	t.Run("should return error when name is empty string", func(t *testing.T) {
		usecase, _ := NewUsecase(t)

		_, err := usecase.Execute(&updatecontest.Input{ID: 1, Name: ptr("")})

		assert.ErrorIs(t, err, updatecontest.ErrorNameIsRequired)
	})

	t.Run("should return error when activate time is zero", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		zeroTime := time.Time{}

		_, err := usecase.Execute(&updatecontest.Input{ID: 1, ActivateTime: &zeroTime})

		assert.ErrorIs(t, err, updatecontest.ErrorActivateTimeIsRequired)
	})

	t.Run("should return error when deactivate time is zero", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		zeroTime := time.Time{}

		_, err := usecase.Execute(&updatecontest.Input{ID: 1, DeactivateTime: &zeroTime})

		assert.ErrorIs(t, err, updatecontest.ErrorDeactivateTimeIsRequired)
	})

	t.Run("should return error when contest not found", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		m.repository.On("FindByID", int32(99)).Return(nil, assert.AnError).Once()

		_, err := usecase.Execute(&updatecontest.Input{ID: 99, Name: ptr("New Name")})

		assert.Error(t, err)
	})

	t.Run("should return error when repository save fails", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		c := newContest(t)
		m.repository.On("FindByID", int32(1)).Return(c, nil).Once()
		m.repository.On("Save", testifymock.Anything).Return(assert.AnError).Once()

		_, err := usecase.Execute(&updatecontest.Input{ID: 1, Name: ptr("Updated")})

		assert.Error(t, err)
	})

	t.Run("should update description", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		c := newContest(t)
		m.repository.On("FindByID", int32(1)).Return(c, nil).Once()
		m.repository.On("Save", testifymock.Anything).Return(nil).Once()

		_, err := usecase.Execute(&updatecontest.Input{ID: 1, Description: ptr("New description")})

		assert.NoError(t, err)
	})

	t.Run("should update is_public", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		c := newContest(t)
		m.repository.On("FindByID", int32(1)).Return(c, nil).Once()
		m.repository.On("Save", testifymock.Anything).Return(nil).Once()

		_, err := usecase.Execute(&updatecontest.Input{ID: 1, IsPublic: ptr(false)})

		assert.NoError(t, err)
	})
}
