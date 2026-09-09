package createcontest_test

import (
	createcontest "judge-backend/internal/app/contest_domain/contest/application/usecase/create_contest"
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

func NewUsecase(t *testing.T) (createcontest.Usecase, *UsecaseMocks) {
	m := NewUsecaseMocks(t)
	return createcontest.NewUsecase(m.provider, m.repository), m
}

func validInput() *createcontest.Input {
	return &createcontest.Input{
		Name:           "Test Contest",
		Description:    "A description",
		ActivateTime:   time.Now(),
		DeactivateTime: time.Now().Add(24 * time.Hour),
		IsPublic:       true,
		ProjectConfigs: []createcontest.CreateProjectConfig{},
	}
}

func TestUsecase_Execute(t *testing.T) {
	t.Run("should create contest successfully", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		m.repository.On("Save", testifymock.Anything).Return(nil).Once()

		output, err := usecase.Execute(validInput())

		assert.NoError(t, err)
		assert.NotNil(t, output)
	})

	t.Run("should return error when name is empty", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		input := validInput()
		input.Name = ""

		_, err := usecase.Execute(input)

		assert.ErrorIs(t, err, createcontest.ErrorNameIsRequired)
	})

	t.Run("should return error when activate time is zero", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		input := validInput()
		input.ActivateTime = time.Time{}

		_, err := usecase.Execute(input)

		assert.ErrorIs(t, err, createcontest.ErrorActivateTimeIsRequired)
	})

	t.Run("should return error when deactivate time is zero", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		input := validInput()
		input.DeactivateTime = time.Time{}

		_, err := usecase.Execute(input)

		assert.ErrorIs(t, err, createcontest.ErrorDeactivateTimeIsRequired)
	})

	t.Run("should return error when repository save fails", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		m.repository.On("Save", testifymock.Anything).Return(assert.AnError).Once()

		_, err := usecase.Execute(validInput())

		assert.Error(t, err)
	})

	t.Run("should create contest with project configs using known strategy", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		m.provider.On("HasStrategy", "fifo").Return(true).Once()
		m.repository.On("Save", testifymock.Anything).Return(nil).Once()

		input := validInput()
		input.ProjectConfigs = []createcontest.CreateProjectConfig{
			{
				ProjectID:             1,
				Name:                  "Project 1",
				ActivationTime:        time.Now(),
				DeactivationTime:      time.Now().Add(24 * time.Hour),
				StartTime:             time.Now(),
				EndTime:               time.Now().Add(24 * time.Hour),
				IsPublic:              true,
				JudgePriorityStrategy: "fifo",
			},
		}

		output, err := usecase.Execute(input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
	})

	t.Run("should carry judge duration and scoreboard settings into the saved contest", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		m.provider.On("HasStrategy", "fifo").Return(true).Once()

		var saved *contest.Contest
		m.repository.On("Save", testifymock.Anything).Run(func(args testifymock.Arguments) {
			saved = args.Get(0).(*contest.Contest)
		}).Return(nil).Once()

		duration := int32(45)
		input := validInput()
		input.ProjectConfigs = []createcontest.CreateProjectConfig{
			{
				ProjectID:               1,
				Name:                    "Project 1",
				ActivationTime:          time.Now(),
				DeactivationTime:        time.Now().Add(24 * time.Hour),
				StartTime:               time.Now(),
				EndTime:                 time.Now().Add(24 * time.Hour),
				IsPublic:                true,
				JudgePriorityStrategy:   "fifo",
				MaxJudgeDurationMinutes: &duration,
				ScoreboardCompact:       true,
				ScoreboardMaskNames:     true,
			},
		}

		_, err := usecase.Execute(input)

		assert.NoError(t, err)
		if assert.Len(t, saved.Projects(), 1) {
			p := saved.Projects()[0]
			if assert.NotNil(t, p.MaxJudgeDurationMinutes()) {
				assert.Equal(t, int32(45), *p.MaxJudgeDurationMinutes())
			}
			assert.True(t, p.ScoreboardCompact())
			assert.True(t, p.ScoreboardMaskNames())
		}
	})

	t.Run("should reject a non-positive max judge duration", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		m.provider.On("HasStrategy", "fifo").Return(true).Once()

		duration := int32(0)
		input := validInput()
		input.ProjectConfigs = []createcontest.CreateProjectConfig{
			{
				ProjectID:               1,
				Name:                    "Project 1",
				ActivationTime:          time.Now(),
				DeactivationTime:        time.Now().Add(24 * time.Hour),
				StartTime:               time.Now(),
				EndTime:                 time.Now().Add(24 * time.Hour),
				IsPublic:                true,
				JudgePriorityStrategy:   "fifo",
				MaxJudgeDurationMinutes: &duration,
			},
		}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
	})
}
