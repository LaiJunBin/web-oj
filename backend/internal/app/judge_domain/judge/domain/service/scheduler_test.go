package service_test

import (
	"testing"

	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/gradingitem"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/gradingitem/scoringstrategy"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/judgetype"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/testcase"
	"judge-backend/internal/app/judge_domain/judge/domain/service"

	"github.com/stretchr/testify/assert"
)

func newTestcase(id int32, deps ...int32) *testcase.Testcase {
	return testcase.NewTestcase(&testcase.TestcaseConfig{
		ID:              id,
		Command:         "echo test",
		Score:           10,
		DependTestcases: deps,
	})
}

func newGradingItem(id int32, deps []int32, testcases ...*testcase.Testcase) *gradingitem.GradingItem {
	return gradingitem.NewGradingItem(&gradingitem.GradingItemConfig{
		ID:              id,
		Score:           100,
		ScoringStrategy: scoringstrategy.IncreaseScoreStrategy,
		DependTestcases: deps,
		Testcases:       testcases,
	})
}

func newJudgment(gradingItems ...*gradingitem.GradingItem) *judgment.Judgment {
	return judgment.New(&judgment.JudgmentConfig{
		ContestProjectID: 1,
		SubmissionID:     "sub-1",
		GradingItems:     gradingItems,
		JudgeType:        judgetype.JUDGE,
	})
}

func testcaseIDs(batch *service.Batch) []int32 {
	ids := []int32{}
	for _, tc := range batch.Testcases {
		ids = append(ids, tc.ID())
	}
	return ids
}

func TestNextBatch_NoGradingItems(t *testing.T) {
	j := newJudgment()
	state := service.NewExecutionState()

	assert.Nil(t, service.NextBatch(j, state))
}

func TestNextBatch_AllIndependent(t *testing.T) {
	j := newJudgment(
		newGradingItem(1, nil, newTestcase(1), newTestcase(2)),
		newGradingItem(2, nil, newTestcase(3)),
	)
	state := service.NewExecutionState()

	batch := service.NextBatch(j, state)
	assert.NotNil(t, batch)
	assert.Equal(t, []int32{1, 2, 3}, testcaseIDs(batch))

	// everything dispatched; nothing left
	assert.Nil(t, service.NextBatch(j, state))
}

func TestNextBatch_GradingItemDependencyGatesSecondWave(t *testing.T) {
	j := newJudgment(
		newGradingItem(1, nil, newTestcase(1)),
		newGradingItem(2, []int32{1}, newTestcase(2)),
	)
	state := service.NewExecutionState()

	first := service.NextBatch(j, state)
	assert.Equal(t, []int32{1}, testcaseIDs(first))

	// dependency not met yet: grading item 2 stays closed
	assert.Nil(t, service.NextBatch(j, state))

	state.MarkPassed(1)
	second := service.NextBatch(j, state)
	assert.Equal(t, []int32{2}, testcaseIDs(second))
}

func TestNextBatch_FailedDependencyNeverOpens(t *testing.T) {
	j := newJudgment(
		newGradingItem(1, nil, newTestcase(1)),
		newGradingItem(2, []int32{1}, newTestcase(2)),
	)
	state := service.NewExecutionState()

	first := service.NextBatch(j, state)
	assert.Equal(t, []int32{1}, testcaseIDs(first))

	// testcase 1 did not pass → grading item 2 never becomes executable
	assert.Nil(t, service.NextBatch(j, state))
	assert.Nil(t, service.NextBatch(j, state))
}

func TestNextBatch_TestcaseDependencyWithinOpenedItems(t *testing.T) {
	j := newJudgment(
		newGradingItem(1, nil, newTestcase(1)),
		newGradingItem(2, []int32{1}, newTestcase(2, 1), newTestcase(3, 999)),
	)
	state := service.NewExecutionState()

	first := service.NextBatch(j, state)
	assert.Equal(t, []int32{1}, testcaseIDs(first))
	state.MarkPassed(1)

	// grading item 2 opens; testcase 2's dep (1) is met, testcase 3's dep (999) is not
	second := service.NextBatch(j, state)
	assert.Equal(t, []int32{2}, testcaseIDs(second))
}

func TestNextBatch_GradingItemForResolvesOwner(t *testing.T) {
	gi1 := newGradingItem(1, nil, newTestcase(1))
	gi2 := newGradingItem(2, nil, newTestcase(2))
	j := newJudgment(gi1, gi2)
	state := service.NewExecutionState()

	batch := service.NextBatch(j, state)
	assert.Same(t, gi1, batch.GradingItemFor(1))
	assert.Same(t, gi2, batch.GradingItemFor(2))
}

// Documents the behavior-preserving quirk noted on NextBatch: a testcase whose
// dependency is met only after its grading item was already opened is never
// revisited. If this is ever "fixed", this test must be updated consciously.
func TestNextBatch_UnmetTestcaseInOpenedItemIsNotRevisited(t *testing.T) {
	j := newJudgment(
		newGradingItem(1, nil, newTestcase(1), newTestcase(2, 1)),
	)
	state := service.NewExecutionState()

	first := service.NextBatch(j, state)
	assert.Equal(t, []int32{1}, testcaseIDs(first))
	state.MarkPassed(1)

	// grading item 1 is already opened, so testcase 2 is never dispatched
	assert.Nil(t, service.NextBatch(j, state))
}
