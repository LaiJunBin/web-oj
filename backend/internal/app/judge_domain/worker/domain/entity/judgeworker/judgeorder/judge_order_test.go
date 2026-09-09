package judgeorder_test

import (
	"judge-backend/internal/app/judge_domain/worker/domain/entity/judgeworker/judgeorder"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("valid JUDGE", func(t *testing.T) {
		order, err := judgeorder.New("JUDGE")

		assert.NoError(t, err)
		assert.NotNil(t, order)
		assert.Equal(t, "JUDGE", order.String())
	})

	t.Run("valid JUDGE_PARTIAL", func(t *testing.T) {
		order, err := judgeorder.New("JUDGE_PARTIAL")

		assert.NoError(t, err)
		assert.NotNil(t, order)
		assert.Equal(t, "JUDGE_PARTIAL", order.String())
	})

	t.Run("valid REJUDGE", func(t *testing.T) {
		order, err := judgeorder.New("REJUDGE")

		assert.NoError(t, err)
		assert.NotNil(t, order)
		assert.Equal(t, "REJUDGE", order.String())
	})

	t.Run("invalid value returns error", func(t *testing.T) {
		order, err := judgeorder.New("INVALID")

		assert.Error(t, err)
		assert.Nil(t, order)
		assert.Contains(t, err.Error(), "INVALID")
	})

	t.Run("empty string returns error", func(t *testing.T) {
		order, err := judgeorder.New("")

		assert.Error(t, err)
		assert.Nil(t, order)
	})
}
