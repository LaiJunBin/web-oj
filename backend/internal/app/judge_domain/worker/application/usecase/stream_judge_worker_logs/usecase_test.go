package streamjudgeworkerlogs_test

import (
	"errors"
	"judge-backend/internal/app/judge_domain/worker/application/query/dto"
	streamjudgeworkerlogs "judge-backend/internal/app/judge_domain/worker/application/usecase/stream_judge_worker_logs"
	"judge-backend/internal/app/judge_domain/worker/mocks"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUsecase_Execute(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockQuery := mocks.NewJudgeWorkerQuery(t)
		usecase := streamjudgeworkerlogs.NewUsecase(mockQuery)

		logs := []*dto.Log{
			{Text: "log line 1", Time: time.Now()},
		}
		worker := &dto.JudgeWorker{
			ID:         "worker-1",
			Status:     "running",
			JudgeOrder: []string{"JUDGE"},
			CreatedAt:  time.Now(),
			Logs:       logs,
		}
		mockQuery.On("GetJudgeWorkerByID", "worker-1").Return(worker, nil).Once()

		output, err := usecase.Execute(&streamjudgeworkerlogs.Input{
			JudgeWorkerID: "worker-1",
		})

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, "worker-1", output.ID())
		assert.Equal(t, logs, output.InitLog())
	})

	t.Run("empty ID returns error", func(t *testing.T) {
		mockQuery := mocks.NewJudgeWorkerQuery(t)
		usecase := streamjudgeworkerlogs.NewUsecase(mockQuery)

		output, err := usecase.Execute(&streamjudgeworkerlogs.Input{
			JudgeWorkerID: "",
		})

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.Equal(t, streamjudgeworkerlogs.ErrorJudgeWorkerIDIsRequired, err)
	})

	t.Run("query error is propagated", func(t *testing.T) {
		mockQuery := mocks.NewJudgeWorkerQuery(t)
		usecase := streamjudgeworkerlogs.NewUsecase(mockQuery)

		mockQuery.On("GetJudgeWorkerByID", "worker-999").Return(nil, errors.New("not found")).Once()

		output, err := usecase.Execute(&streamjudgeworkerlogs.Input{
			JudgeWorkerID: "worker-999",
		})

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.EqualError(t, err, "not found")
	})

	t.Run("output InitLog returns nil when no logs", func(t *testing.T) {
		mockQuery := mocks.NewJudgeWorkerQuery(t)
		usecase := streamjudgeworkerlogs.NewUsecase(mockQuery)

		worker := &dto.JudgeWorker{
			ID:         "worker-1",
			Status:     "running",
			JudgeOrder: []string{"JUDGE"},
			CreatedAt:  time.Now(),
			Logs:       nil,
		}
		mockQuery.On("GetJudgeWorkerByID", "worker-1").Return(worker, nil).Once()

		output, err := usecase.Execute(&streamjudgeworkerlogs.Input{
			JudgeWorkerID: "worker-1",
		})

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Nil(t, output.InitLog())
	})
}
