package stopjudgeworker_test

import (
	"errors"
	stopjudgeworker "judge-backend/internal/app/judge_domain/worker/application/usecase/stop_judge_worker"
	"judge-backend/internal/app/judge_domain/worker/domain/entity/judgeworker"
	"judge-backend/internal/app/judge_domain/worker/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUsecase_Execute(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockRepo := mocks.NewJudgeWorkerRepository(t)
		mockService := mocks.NewJudgeWorkerService(t)
		usecase := stopjudgeworker.NewUsecase(mockRepo, mockService)

		worker, _ := judgeworker.New([]string{"JUDGE"})
		mockRepo.On("FindByID", "worker-1").Return(worker, nil).Once()
		mockService.On("Stop", worker).Return(nil).Once()

		output, err := usecase.Execute(&stopjudgeworker.Input{
			JudgeWorkerID: "worker-1",
		})

		assert.NoError(t, err)
		assert.NotNil(t, output)
	})

	t.Run("empty ID returns error", func(t *testing.T) {
		mockRepo := mocks.NewJudgeWorkerRepository(t)
		mockService := mocks.NewJudgeWorkerService(t)
		usecase := stopjudgeworker.NewUsecase(mockRepo, mockService)

		output, err := usecase.Execute(&stopjudgeworker.Input{
			JudgeWorkerID: "",
		})

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.Equal(t, stopjudgeworker.ErrorJudgeWorkerIDIsRequired, err)
	})

	t.Run("FindByID error is propagated", func(t *testing.T) {
		mockRepo := mocks.NewJudgeWorkerRepository(t)
		mockService := mocks.NewJudgeWorkerService(t)
		usecase := stopjudgeworker.NewUsecase(mockRepo, mockService)

		mockRepo.On("FindByID", "worker-999").Return(nil, errors.New("not found")).Once()

		output, err := usecase.Execute(&stopjudgeworker.Input{
			JudgeWorkerID: "worker-999",
		})

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.EqualError(t, err, "not found")
	})

	t.Run("service Stop error is propagated", func(t *testing.T) {
		mockRepo := mocks.NewJudgeWorkerRepository(t)
		mockService := mocks.NewJudgeWorkerService(t)
		usecase := stopjudgeworker.NewUsecase(mockRepo, mockService)

		worker, _ := judgeworker.New([]string{"JUDGE"})
		mockRepo.On("FindByID", "worker-1").Return(worker, nil).Once()
		mockService.On("Stop", worker).Return(errors.New("stop failed")).Once()

		output, err := usecase.Execute(&stopjudgeworker.Input{
			JudgeWorkerID: "worker-1",
		})

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.EqualError(t, err, "stop failed")
	})
}
