package getjudgequeue_test

import (
	"testing"

	"judge-backend/internal/app/judge_domain/worker/application/query/dto"
	getjudgequeue "judge-backend/internal/app/judge_domain/worker/application/usecase/get_judge_queue"
	"judge-backend/internal/app/judge_domain/worker/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsecase_Execute(t *testing.T) {
	t.Run("returns the queue status", func(t *testing.T) {
		q := mocks.NewJudgeQueueQuery(t)
		uc := getjudgequeue.NewUsecase(q)

		waited := 12.5
		q.On("GetJudgeQueueStatus").Return(&dto.JudgeQueueStatus{
			QueueDepths:         map[string]int32{"JUDGE": 2, "JUDGE_PARTIAL": 0, "REJUDGE": 1},
			ActiveJudgments:     3,
			OldestQueuedSeconds: &waited,
		}, nil).Once()

		output, err := uc.Execute(&getjudgequeue.Input{})

		require.NoError(t, err)
		assert.EqualValues(t, 3, output.Status.ActiveJudgments)
		require.NotNil(t, output.Status.OldestQueuedSeconds)
		assert.InDelta(t, 12.5, *output.Status.OldestQueuedSeconds, 0.001)
	})

	t.Run("returns the error when the query fails", func(t *testing.T) {
		q := mocks.NewJudgeQueueQuery(t)
		uc := getjudgequeue.NewUsecase(q)

		q.On("GetJudgeQueueStatus").Return((*dto.JudgeQueueStatus)(nil), assert.AnError).Once()

		_, err := uc.Execute(&getjudgequeue.Input{})

		assert.Error(t, err)
	})
}
