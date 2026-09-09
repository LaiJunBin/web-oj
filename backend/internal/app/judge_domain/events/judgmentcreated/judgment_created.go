package judgmentcreated

import (
	"fmt"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment"

	"github.com/google/uuid"
)

type JudgmentCreatedEvent struct {
	uuid     string
	judgment *judgment.Judgment
}

func (e *JudgmentCreatedEvent) Topic() string {
	return "judgment_created"
}

func NewJudgmentCreatedEvent(judgment *judgment.Judgment) *JudgmentCreatedEvent {
	return &JudgmentCreatedEvent{
		uuid:     uuid.New().String(),
		judgment: judgment,
	}
}

func (e *JudgmentCreatedEvent) ID() string {
	return fmt.Sprintf("%s-%s-%s", e.Topic(), e.judgment.ID(), e.uuid)
}

func (e *JudgmentCreatedEvent) Judgment() *judgment.Judgment {
	return e.judgment
}
