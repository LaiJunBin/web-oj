package deletejudgeworker_test

import (
	"errors"
	deletejudgeworker "judge-backend/internal/app/judge_domain/worker/application/usecase/delete_judge_worker"
	"judge-backend/internal/app/judge_domain/worker/domain/entity/judgeworker"
	"judge-backend/internal/app/judge_domain/worker/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUsecase_Execute(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockRepo := mocks.NewJudgeWorkerRepository(t)
		usecase := deletejudgeworker.NewUsecase(mockRepo)

		worker, _ := judgeworker.New([]string{"JUDGE"})
		mockRepo.On("FindByID", "worker-1").Return(worker, nil).Once()
		mockRepo.On("Delete", worker).Return(nil).Once()

		output, err := usecase.Execute(&deletejudgeworker.Input{
			JudgeWorkerID: "worker-1",
		})

		assert.NoError(t, err)
		assert.NotNil(t, output)
	})

	t.Run("empty ID returns error", func(t *testing.T) {
		mockRepo := mocks.NewJudgeWorkerRepository(t)
		usecase := deletejudgeworker.NewUsecase(mockRepo)

		output, err := usecase.Execute(&deletejudgeworker.Input{
			JudgeWorkerID: "",
		})

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.Equal(t, deletejudgeworker.ErrorJudgeWorkerIDIsRequired, err)
	})

	t.Run("FindByID error is wrapped with 400", func(t *testing.T) {
		mockRepo := mocks.NewJudgeWorkerRepository(t)
		usecase := deletejudgeworker.NewUsecase(mockRepo)

		mockRepo.On("FindByID", "worker-999").Return(nil, errors.New("not found")).Once()

		output, err := usecase.Execute(&deletejudgeworker.Input{
			JudgeWorkerID: "worker-999",
		})

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("Delete error is propagated", func(t *testing.T) {
		mockRepo := mocks.NewJudgeWorkerRepository(t)
		usecase := deletejudgeworker.NewUsecase(mockRepo)

		worker, _ := judgeworker.New([]string{"JUDGE"})
		mockRepo.On("FindByID", "worker-1").Return(worker, nil).Once()
		mockRepo.On("Delete", worker).Return(errors.New("delete failed")).Once()

		output, err := usecase.Execute(&deletejudgeworker.Input{
			JudgeWorkerID: "worker-1",
		})

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.EqualError(t, err, "delete failed")
	})
}
