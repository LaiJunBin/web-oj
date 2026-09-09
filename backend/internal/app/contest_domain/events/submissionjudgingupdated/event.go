package submissionjudgingupdated

import "judge-backend/internal/app/contest_domain/events/types"

type Event struct {
	id               string
	contestProjectID int32
	judgment         types.JudgmentData
}

func New(id string, contestProjectID int32, judgment types.JudgmentData) *Event {
	return &Event{
		id:               id,
		contestProjectID: contestProjectID,
		judgment:         judgment,
	}
}

func (e *Event) Topic() string                { return "contest_submission_judging_updated" }
func (e *Event) ID() string                   { return e.id }
func (e *Event) ContestProjectID() int32      { return e.contestProjectID }
func (e *Event) Judgment() types.JudgmentData { return e.judgment }
