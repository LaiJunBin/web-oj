package listjudgeprioritystrategies_test

import (
	"errors"
	listjudgeprioritystrategies "judge-backend/internal/app/judge_domain/worker/application/usecase/list_judge_priority_strategies"
	"judge-backend/internal/app/judge_domain/worker/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUsecase_Execute(t *testing.T) {
	t.Run("success with strategies", func(t *testing.T) {
		mockQuery := mocks.NewJudgeWorkerQuery(t)
		usecase := listjudgeprioritystrategies.NewUsecase(mockQuery)

		expected := []string{"default", "fifo", "priority"}
		mockQuery.On("ListJudgePriorityStrategies").Return(expected, nil).Once()

		output, err := usecase.Execute(&listjudgeprioritystrategies.Input{})

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, expected, output.JudgePriorityStrategies)
	})

	t.Run("success with empty list", func(t *testing.T) {
		mockQuery := mocks.NewJudgeWorkerQuery(t)
		usecase := listjudgeprioritystrategies.NewUsecase(mockQuery)

		mockQuery.On("ListJudgePriorityStrategies").Return([]string{}, nil).Once()

		output, err := usecase.Execute(&listjudgeprioritystrategies.Input{})

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Empty(t, output.JudgePriorityStrategies)
	})

	t.Run("query error is propagated", func(t *testing.T) {
		mockQuery := mocks.NewJudgeWorkerQuery(t)
		usecase := listjudgeprioritystrategies.NewUsecase(mockQuery)

		mockQuery.On("ListJudgePriorityStrategies").Return(nil, errors.New("service error")).Once()

		output, err := usecase.Execute(&listjudgeprioritystrategies.Input{})

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.EqualError(t, err, "service error")
	})
}
