package canceljudgerequested

type Event struct {
	id         string
	judgmentID string
	reason     string
}

func New(id, judgmentID, reason string) *Event {
	return &Event{id: id, judgmentID: judgmentID, reason: reason}
}

func (e *Event) Topic() string      { return "judge_cancel_judge_requested" }
func (e *Event) ID() string         { return e.id }
func (e *Event) JudgmentID() string { return e.judgmentID }
func (e *Event) Reason() string     { return e.reason }
