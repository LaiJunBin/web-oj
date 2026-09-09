package addjudgeworker_test

import (
	"errors"
	addjudgeworker "judge-backend/internal/app/judge_domain/worker/application/usecase/add_judge_worker"
	"judge-backend/internal/app/judge_domain/worker/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUsecase_Execute(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockRepo := mocks.NewJudgeWorkerRepository(t)
		usecase := addjudgeworker.NewUsecase(mockRepo)

		mockRepo.On("Save", mock.Anything).Return(nil).Once()

		output, err := usecase.Execute(&addjudgeworker.Input{
			JudgeOrder: []string{"JUDGE", "REJUDGE"},
		})

		assert.NoError(t, err)
		assert.NotNil(t, output)
	})

	t.Run("nil JudgeOrder returns error", func(t *testing.T) {
		mockRepo := mocks.NewJudgeWorkerRepository(t)
		usecase := addjudgeworker.NewUsecase(mockRepo)

		output, err := usecase.Execute(&addjudgeworker.Input{
			JudgeOrder: nil,
		})

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.Equal(t, addjudgeworker.ErrorJudgeOrderIsRequired, err)
	})

	t.Run("empty JudgeOrder returns error", func(t *testing.T) {
		mockRepo := mocks.NewJudgeWorkerRepository(t)
		usecase := addjudgeworker.NewUsecase(mockRepo)

		output, err := usecase.Execute(&addjudgeworker.Input{
			JudgeOrder: []string{},
		})

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.Equal(t, addjudgeworker.ErrorJudgeOrderIsRequired, err)
	})

	t.Run("invalid judge order value returns error", func(t *testing.T) {
		mockRepo := mocks.NewJudgeWorkerRepository(t)
		usecase := addjudgeworker.NewUsecase(mockRepo)

		output, err := usecase.Execute(&addjudgeworker.Input{
			JudgeOrder: []string{"INVALID_ORDER"},
		})

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.Contains(t, err.Error(), "INVALID_ORDER")
	})

	t.Run("repo Save error is propagated", func(t *testing.T) {
		mockRepo := mocks.NewJudgeWorkerRepository(t)
		usecase := addjudgeworker.NewUsecase(mockRepo)

		mockRepo.On("Save", mock.Anything).Return(errors.New("db error")).Once()

		output, err := usecase.Execute(&addjudgeworker.Input{
			JudgeOrder: []string{"JUDGE"},
		})

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.EqualError(t, err, "db error")
	})
}
