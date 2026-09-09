package submissionupdated

import "judge-backend/internal/app/contest_domain/events/types"

type Event struct {
	contestProjectID int32
	judgment         types.JudgmentData
}

func New(contestProjectID int32, judgment types.JudgmentData) *Event {
	return &Event{
		contestProjectID: contestProjectID,
		judgment:         judgment,
	}
}

func (e *Event) Topic() string                { return "submission_updated" }
func (e *Event) ContestProjectID() int32      { return e.contestProjectID }
func (e *Event) Judgment() types.JudgmentData { return e.judgment }
