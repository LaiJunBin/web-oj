package rejudgerequested

type Event struct {
	id               string
	contestProjectID int32
	submissionIDs    []string
	reason           string
}

func New(id string, contestProjectID int32, submissionIDs []string, reason string) *Event {
	return &Event{id: id, contestProjectID: contestProjectID, submissionIDs: submissionIDs, reason: reason}
}

func (e *Event) Topic() string           { return "judge_rejudge_requested" }
func (e *Event) ID() string              { return e.id }
func (e *Event) ContestProjectID() int32 { return e.contestProjectID }
func (e *Event) SubmissionIDs() []string { return e.submissionIDs }
func (e *Event) Reason() string          { return e.reason }
