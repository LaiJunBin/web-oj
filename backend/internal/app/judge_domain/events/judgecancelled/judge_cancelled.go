package judgecancelled

type JudgeCancelledEvent struct {
	judgmentID string
}

func (e *JudgeCancelledEvent) Topic() string {
	return "judge_cancelled"
}

func NewJudgeCancelledEvent(
	judgmentID string,
) *JudgeCancelledEvent {
	return &JudgeCancelledEvent{
		judgmentID: judgmentID,
	}
}

func (e *JudgeCancelledEvent) ID() string {
	return e.judgmentID
}

func (e *JudgeCancelledEvent) JudgmentID() string {
	return e.judgmentID
}
