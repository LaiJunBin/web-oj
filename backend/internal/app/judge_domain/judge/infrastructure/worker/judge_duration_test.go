package worker

import (
	"testing"
	"time"

	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/judgetype"

	"github.com/stretchr/testify/assert"
)

func durationTestJudgment(t *testing.T, minutes *int32) *judgment.Judgment {
	t.Helper()
	j := judgment.New(&judgment.JudgmentConfig{
		ContestProjectID: 1,
		SubmissionID:     "sub-1",
		JudgeType:        judgetype.JUDGE,
	})
	j.SetMaxJudgeDurationMinutes(minutes)
	return j
}

func int32Ptr(v int32) *int32 { return &v }

func TestJudgeDurationFor(t *testing.T) {
	fallback := 30 * time.Minute

	t.Run("no override uses the global default", func(t *testing.T) {
		assert.Equal(t, fallback, judgeDurationFor(durationTestJudgment(t, nil), fallback))
	})

	t.Run("override wins over the global default", func(t *testing.T) {
		assert.Equal(t, 45*time.Minute, judgeDurationFor(durationTestJudgment(t, int32Ptr(45)), fallback))
	})

	t.Run("non-positive override is ignored", func(t *testing.T) {
		assert.Equal(t, fallback, judgeDurationFor(durationTestJudgment(t, int32Ptr(0)), fallback))
		assert.Equal(t, fallback, judgeDurationFor(durationTestJudgment(t, int32Ptr(-5)), fallback))
	})
}
