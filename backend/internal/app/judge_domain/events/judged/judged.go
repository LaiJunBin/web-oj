package judged

import "judge-backend/internal/app/judge_domain/judge/domain/entity/judgment"

type JudgedEvent struct {
	id               string
	judgmentID       string
	contestProjectID int32
}

func (e *JudgedEvent) Topic() string {
	return "judged"
}

func NewJudgedEvent(j *judgment.Judgment) *JudgedEvent {
	return &JudgedEvent{
		id:               j.ID(),
		judgmentID:       j.ID(),
		contestProjectID: j.ContestProjectID(),
	}
}

func (e *JudgedEvent) ID() string              { return e.id }
func (e *JudgedEvent) JudgmentID() string      { return e.judgmentID }
func (e *JudgedEvent) ContestProjectID() int32 { return e.contestProjectID }
