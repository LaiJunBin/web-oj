package getjudgeworker_test

import (
	"errors"
	"judge-backend/internal/app/judge_domain/worker/application/query/dto"
	getjudgeworker "judge-backend/internal/app/judge_domain/worker/application/usecase/get_judge_worker"
	"judge-backend/internal/app/judge_domain/worker/mocks"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUsecase_Execute(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockQuery := mocks.NewJudgeWorkerQuery(t)
		usecase := getjudgeworker.NewUsecase(mockQuery)

		expected := &dto.JudgeWorker{
			ID:         "worker-1",
			Status:     "running",
			JudgeOrder: []string{"JUDGE"},
			CreatedAt:  time.Now(),
			Logs:       []*dto.Log{},
		}
		mockQuery.On("GetJudgeWorkerByID", "worker-1").Return(expected, nil).Once()

		output, err := usecase.Execute(&getjudgeworker.Input{
			JudgeWorkerID: "worker-1",
		})

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, expected, output.JudgeWorker)
	})

	t.Run("empty ID returns error", func(t *testing.T) {
		mockQuery := mocks.NewJudgeWorkerQuery(t)
		usecase := getjudgeworker.NewUsecase(mockQuery)

		output, err := usecase.Execute(&getjudgeworker.Input{
			JudgeWorkerID: "",
		})

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.Equal(t, getjudgeworker.ErrorJudgeWorkerIDIsRequired, err)
	})

	t.Run("query error is propagated", func(t *testing.T) {
		mockQuery := mocks.NewJudgeWorkerQuery(t)
		usecase := getjudgeworker.NewUsecase(mockQuery)

		mockQuery.On("GetJudgeWorkerByID", "worker-999").Return(nil, errors.New("not found")).Once()

		output, err := usecase.Execute(&getjudgeworker.Input{
			JudgeWorkerID: "worker-999",
		})

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.EqualError(t, err, "not found")
	})
}
