package submissioncreated

import "fmt"

type SubmissionCreatedEvent struct {
	submissionID     string
	contestProjectID int32
	isPartial        bool
}

func (s *SubmissionCreatedEvent) Topic() string {
	return "submission_created"
}

func NewSubmissionCreatedEvent(submissionID string, contestProjectID int32, isPartial bool) *SubmissionCreatedEvent {
	return &SubmissionCreatedEvent{
		submissionID:     submissionID,
		contestProjectID: contestProjectID,
		isPartial:        isPartial,
	}
}

func (e *SubmissionCreatedEvent) ID() string {
	return fmt.Sprintf("%s-%s", e.Topic(), e.submissionID)
}

func (s *SubmissionCreatedEvent) SubmissionID() string {
	return s.submissionID
}

func (s *SubmissionCreatedEvent) ContestProjectID() int32 {
	return s.contestProjectID
}

func (s *SubmissionCreatedEvent) IsPartial() bool {
	return s.isPartial
}
