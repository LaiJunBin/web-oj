package deletecontest_test

import (
	deletecontest "judge-backend/internal/app/contest_domain/contest/application/usecase/delete_contest"
	"judge-backend/internal/app/contest_domain/contest/domain/entity/contest"
	"judge-backend/internal/app/contest_domain/contest/mocks"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	testifymock "github.com/stretchr/testify/mock"
)

type UsecaseMocks struct {
	repository *mocks.ContestRepository
}

func NewUsecaseMocks(t *testing.T) *UsecaseMocks {
	return &UsecaseMocks{
		repository: mocks.NewContestRepository(t),
	}
}

func NewUsecase(t *testing.T) (deletecontest.Usecase, *UsecaseMocks) {
	m := NewUsecaseMocks(t)
	return deletecontest.NewUsecase(m.repository), m
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

func TestUsecase_Execute(t *testing.T) {
	t.Run("should delete contest successfully", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		c := newContest(t)
		m.repository.On("FindByID", int32(1)).Return(c, nil).Once()
		m.repository.On("Delete", testifymock.Anything).Return(nil).Once()

		_, err := usecase.Execute(&deletecontest.Input{ID: 1})

		assert.NoError(t, err)
	})

	t.Run("should return error when id is invalid", func(t *testing.T) {
		usecase, _ := NewUsecase(t)

		_, err := usecase.Execute(&deletecontest.Input{ID: 0})

		assert.ErrorIs(t, err, deletecontest.ErrorInvalidID)
	})

	t.Run("should return error when contest not found", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		m.repository.On("FindByID", int32(99)).Return(nil, assert.AnError).Once()

		_, err := usecase.Execute(&deletecontest.Input{ID: 99})

		assert.Error(t, err)
	})

	t.Run("should return error when repository delete fails", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		c := newContest(t)
		m.repository.On("FindByID", int32(1)).Return(c, nil).Once()
		m.repository.On("Delete", testifymock.Anything).Return(assert.AnError).Once()

		_, err := usecase.Execute(&deletecontest.Input{ID: 1})

		assert.Error(t, err)
	})
}
