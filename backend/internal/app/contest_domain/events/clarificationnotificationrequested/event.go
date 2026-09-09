package clarificationnotificationrequested

import "judge-backend/internal/app/contest_domain/activity/domain/entity/clarification"

type Event struct {
	clarification *clarification.Clarification
}

func New(c *clarification.Clarification) *Event {
	return &Event{clarification: c}
}

func (e *Event) Topic() string { return "clarification_notification_requested" }
func (e *Event) Clarification() *clarification.Clarification {
	return e.clarification
}
