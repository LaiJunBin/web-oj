package listjudgeworker_test

import (
	"errors"
	"judge-backend/internal/app/judge_domain/worker/application/query/dto"
	listjudgeworker "judge-backend/internal/app/judge_domain/worker/application/usecase/list_judge_worker"
	"judge-backend/internal/app/judge_domain/worker/mocks"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUsecase_Execute(t *testing.T) {
	t.Run("success with workers", func(t *testing.T) {
		mockQuery := mocks.NewJudgeWorkerQuery(t)
		usecase := listjudgeworker.NewUsecase(mockQuery)

		expected := []*dto.JudgeWorkerInfo{
			{
				ID:         "worker-1",
				Status:     "running",
				JudgeOrder: []string{"JUDGE"},
				CreatedAt:  time.Now(),
			},
		}
		mockQuery.On("ListJudgeWorker").Return(expected, nil).Once()

		output, err := usecase.Execute(&listjudgeworker.Input{})

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, expected, output.JudgeWorkers)
	})

	t.Run("success with empty list", func(t *testing.T) {
		mockQuery := mocks.NewJudgeWorkerQuery(t)
		usecase := listjudgeworker.NewUsecase(mockQuery)

		mockQuery.On("ListJudgeWorker").Return([]*dto.JudgeWorkerInfo{}, nil).Once()

		output, err := usecase.Execute(&listjudgeworker.Input{})

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Empty(t, output.JudgeWorkers)
	})

	t.Run("query error is propagated", func(t *testing.T) {
		mockQuery := mocks.NewJudgeWorkerQuery(t)
		usecase := listjudgeworker.NewUsecase(mockQuery)

		mockQuery.On("ListJudgeWorker").Return(nil, errors.New("db error")).Once()

		output, err := usecase.Execute(&listjudgeworker.Input{})

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.EqualError(t, err, "db error")
	})
}
