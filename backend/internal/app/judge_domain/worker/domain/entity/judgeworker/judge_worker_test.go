package judgeworker_test

import (
	"judge-backend/internal/app/judge_domain/worker/domain/entity/judgeworker"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("success with valid judge orders", func(t *testing.T) {
		worker, err := judgeworker.New([]string{"JUDGE", "REJUDGE"})

		assert.NoError(t, err)
		assert.NotNil(t, worker)
		assert.Equal(t, "", worker.ID())
		orders := worker.JudgeOrder()
		assert.Len(t, orders, 2)
		assert.Equal(t, "JUDGE", orders[0].String())
		assert.Equal(t, "REJUDGE", orders[1].String())
	})

	t.Run("success with empty judge orders", func(t *testing.T) {
		worker, err := judgeworker.New([]string{})

		assert.NoError(t, err)
		assert.NotNil(t, worker)
		assert.Empty(t, worker.JudgeOrder())
	})

	t.Run("returns error for invalid judge order", func(t *testing.T) {
		worker, err := judgeworker.New([]string{"JUDGE", "INVALID"})

		assert.Error(t, err)
		assert.Nil(t, worker)
	})
}

func TestSetID(t *testing.T) {
	worker, err := judgeworker.New([]string{"JUDGE"})
	assert.NoError(t, err)

	assert.Equal(t, "", worker.ID())
	worker.SetID("worker-123")
	assert.Equal(t, "worker-123", worker.ID())
}
