package prioritystrategies

import (
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment"
	"judge-backend/internal/app/judge_domain/judge/domain/query"
	"time"
)

const (
	prevSubmissionTimeField = "prev_submission_time"
	hasPrevSubmissionField  = "has_prev_submission"
)

type TimeSinceLast struct {
	query query.JudgmentQuery
}

func NewTimeSinceLast(
	query query.JudgmentQuery,
) Strategy {
	return &TimeSinceLast{
		query: query,
	}
}

func (s *TimeSinceLast) Name() string {
	return "time_since_last"
}

func (s *TimeSinceLast) Priority(i, j *judgment.Judgment) bool {
	iHasPrev := hasPreviousSubmission(i)
	jHasPrev := hasPreviousSubmission(j)

	if !iHasPrev && jHasPrev {
		return true
	}
	if iHasPrev && !jHasPrev {
		return false
	}

	if !iHasPrev && !jHasPrev {
		return i.CreatedAt().Before(j.CreatedAt())
	}

	t1, ok := previousSubmissionTime(i)
	if !ok {
		return i.CreatedAt().Before(j.CreatedAt())
	}
	t2, ok := previousSubmissionTime(j)
	if !ok {
		return i.CreatedAt().Before(j.CreatedAt())
	}

	if t1.Equal(t2) {
		return i.CreatedAt().Before(j.CreatedAt())
	}
	return t1.Before(t2)
}

func (s *TimeSinceLast) GetExtraFields(j *judgment.Judgment) map[string]interface{} {
	m := map[string]interface{}{
		hasPrevSubmissionField:  false,
		prevSubmissionTimeField: time.Time{},
	}

	t, err := s.query.GetPrevSubmissionTimeByJudgmentID(j.ID())
	if err != nil || t.IsZero() {
		return m
	}

	m[hasPrevSubmissionField] = true
	m[prevSubmissionTimeField] = t
	return m
}

func hasPreviousSubmission(j *judgment.Judgment) bool {
	v, ok := j.SortingFields()[hasPrevSubmissionField]
	if !ok {
		return false
	}
	hasPrev, ok := v.(bool)
	return ok && hasPrev
}

func previousSubmissionTime(j *judgment.Judgment) (time.Time, bool) {
	v, ok := j.SortingFields()[prevSubmissionTimeField]
	if !ok {
		return time.Time{}, false
	}
	t, ok := v.(time.Time)
	if !ok || t.IsZero() {
		return time.Time{}, false
	}
	return t, true
}
