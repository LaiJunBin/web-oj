package rejudgeapprovalrequested

type Event struct {
	id        string
	rejudgeID string
}

func New(id, rejudgeID string) *Event {
	return &Event{id: id, rejudgeID: rejudgeID}
}

func (e *Event) Topic() string     { return "judge_rejudge_approval_requested" }
func (e *Event) ID() string        { return e.id }
func (e *Event) RejudgeID() string { return e.rejudgeID }
