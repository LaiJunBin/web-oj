package prioritystrategies

import (
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment"
)

type SubmitTime struct{}

func NewSubmitTime() Strategy {
	return &SubmitTime{}
}

func (s *SubmitTime) Name() string {
	return "submit_time"
}

func (s *SubmitTime) Priority(i, j *judgment.Judgment) bool {
	return i.CreatedAt().Before(j.CreatedAt())
}

func (s *SubmitTime) GetExtraFields(j *judgment.Judgment) map[string]interface{} {
	return nil
}
