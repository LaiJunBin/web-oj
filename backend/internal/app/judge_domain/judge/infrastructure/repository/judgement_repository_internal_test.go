package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// validJudgment returns a minimal Judgment bson model that toDomainModel can
// convert successfully, as a baseline for the mutation-based tests below.
func validJudgment() *Judgment {
	j := &Judgment{
		ID:        "judgment-1",
		Status:    "Queued",
		JudgeType: "JUDGE",
	}
	j.GradingItems = append(j.GradingItems, judgmentGradingItem{
		ID:              1,
		ScoringStrategy: "increase",
		Testcases: []judgmentTestcase{
			{ID: 1, Command: "echo hi", Status: "Pending"},
		},
	})
	return j
}

// TestToDomainModel_UnknownEnums is a regression test for A4: toDomainModel
// used to return (nil, nil) — success with a nil judgment — whenever a
// stored document contained an enum string it didn't recognize (status,
// judgeType, testcase status, scoringStrategy). Every caller checks `err !=
// nil` before using the result, so a bare nil slipped through as if the
// lookup had succeeded, crashing the first caller that dereferenced it.
func TestToDomainModel_UnknownEnums(t *testing.T) {
	r := &JudgmentRepository{}

	t.Run("valid document converts without error", func(t *testing.T) {
		model, err := r.toDomainModel(validJudgment())
		assert.NoError(t, err)
		assert.NotNil(t, model)
	})

	t.Run("unknown judgment status returns an error, not (nil, nil)", func(t *testing.T) {
		j := validJudgment()
		j.Status = "SomeFutureStatus"
		model, err := r.toDomainModel(j)
		assert.Error(t, err)
		assert.Nil(t, model)
		assert.ErrorContains(t, err, "SomeFutureStatus")
	})

	t.Run("unknown judgeType returns an error, not (nil, nil)", func(t *testing.T) {
		j := validJudgment()
		j.JudgeType = "SOME_FUTURE_TYPE"
		model, err := r.toDomainModel(j)
		assert.Error(t, err)
		assert.Nil(t, model)
		assert.ErrorContains(t, err, "SOME_FUTURE_TYPE")
	})

	t.Run("unknown testcase status returns an error, not (nil, nil)", func(t *testing.T) {
		j := validJudgment()
		j.GradingItems[0].Testcases[0].Status = "SomeFutureTestcaseStatus"
		model, err := r.toDomainModel(j)
		assert.Error(t, err)
		assert.Nil(t, model)
		assert.ErrorContains(t, err, "SomeFutureTestcaseStatus")
	})

	t.Run("unknown scoringStrategy returns an error, not (nil, nil)", func(t *testing.T) {
		j := validJudgment()
		j.GradingItems[0].ScoringStrategy = "some_future_strategy"
		model, err := r.toDomainModel(j)
		assert.Error(t, err)
		assert.Nil(t, model)
		assert.ErrorContains(t, err, "some_future_strategy")
	})
}
