package prioritystrategies

import (
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment"
)

type Default struct {
	strategy Strategy
}

func NewDefault() Strategy {
	return &Default{
		strategy: NewSubmitTime(),
	}
}

func (d *Default) Name() string {
	return "default"
}

func (d *Default) Priority(i, j *judgment.Judgment) bool {
	return d.strategy.Priority(i, j)
}

func (d *Default) GetExtraFields(j *judgment.Judgment) map[string]interface{} {
	return nil
}
