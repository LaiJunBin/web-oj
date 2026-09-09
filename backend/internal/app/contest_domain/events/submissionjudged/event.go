package submissionjudged

type Event struct {
	id               string
	contestProjectID int32
}

func New(id string, contestProjectID int32) *Event {
	return &Event{
		id:               id,
		contestProjectID: contestProjectID,
	}
}

func (e *Event) Topic() string           { return "contest_submission_judged" }
func (e *Event) ID() string              { return e.id }
func (e *Event) ContestProjectID() int32 { return e.contestProjectID }
