package prioritystrategies_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment"
	judgmentproject "judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/project"
	prioritystrategies "judge-backend/internal/app/judge_domain/judge/domain/prioritystrategies"
	"judge-backend/internal/app/judge_domain/judge/mocks"
)

const (
	prioritystrategiesPrevSubmissionTimeField = "prev_submission_time"
	prioritystrategiesHasPrevSubmissionField  = "has_prev_submission"
)

func newStrategyJudgment(id string, createdAt time.Time) *judgment.Judgment {
	j := judgment.New(&judgment.JudgmentConfig{
		Project:   judgmentproject.NewProject(&judgmentproject.ProjectConfig{ID: 1}),
		CreatedAt: createdAt,
	})
	j.SetID(id)
	return j
}

func TestTimeSinceLast_GetExtraFields_NoPreviousSubmissionIsTreatedAsFirstSubmission(t *testing.T) {
	query := mocks.NewJudgmentQuery(t)
	strategy := prioritystrategies.NewTimeSinceLast(query)
	j := newStrategyJudgment("judgment-1", time.Now())

	query.On("GetPrevSubmissionTimeByJudgmentID", "judgment-1").Return(time.Time{}, nil).Once()

	fields := strategy.GetExtraFields(j)
	assert.Equal(t, false, fields[prioritystrategiesHasPrevSubmissionField])
	assert.Equal(t, time.Time{}, fields[prioritystrategiesPrevSubmissionTimeField])
}

func TestTimeSinceLast_GetExtraFields_QueryErrorFallsBackToFirstSubmissionPriority(t *testing.T) {
	query := mocks.NewJudgmentQuery(t)
	strategy := prioritystrategies.NewTimeSinceLast(query)
	j := newStrategyJudgment("judgment-1", time.Now())

	query.On("GetPrevSubmissionTimeByJudgmentID", "judgment-1").Return(time.Time{}, errors.New("temporary failure")).Once()

	fields := strategy.GetExtraFields(j)
	assert.Equal(t, false, fields[prioritystrategiesHasPrevSubmissionField])
	assert.Equal(t, time.Time{}, fields[prioritystrategiesPrevSubmissionTimeField])
}

func TestTimeSinceLast_Priority_FirstSubmissionBeatsRepeatSubmission(t *testing.T) {
	query := mocks.NewJudgmentQuery(t)
	strategy := prioritystrategies.NewTimeSinceLast(query)
	base := time.Now()

	first := newStrategyJudgment("first", base)
	first.SetSortingFields(map[string]interface{}{
		prioritystrategiesHasPrevSubmissionField:  false,
		prioritystrategiesPrevSubmissionTimeField: time.Time{},
	})
	repeat := newStrategyJudgment("repeat", base.Add(time.Second))
	repeat.SetSortingFields(map[string]interface{}{
		prioritystrategiesHasPrevSubmissionField:  true,
		prioritystrategiesPrevSubmissionTimeField: base.Add(-time.Minute),
	})

	assert.True(t, strategy.Priority(first, repeat))
	assert.False(t, strategy.Priority(repeat, first))
}
