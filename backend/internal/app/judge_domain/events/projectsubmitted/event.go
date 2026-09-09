package projectsubmitted

import (
	"bytes"
	"judge-backend/internal/app/shared/types/submissiondynamicfield"
	"time"
)

type Event struct {
	id               string
	userID           int32
	contestProjectID int32
	isPartial        bool
	gradingItemIds   []int32
	project          *bytes.Buffer
	dynamicFields    []*submissiondynamicfield.SubmissionDynamicField
	createdAt        time.Time
}

func New(
	id string,
	userID int32,
	contestProjectID int32,
	isPartial bool,
	gradingItemIds []int32,
	project *bytes.Buffer,
	dynamicFields []*submissiondynamicfield.SubmissionDynamicField,
	createdAt time.Time,
) *Event {
	return &Event{
		id:               id,
		userID:           userID,
		contestProjectID: contestProjectID,
		isPartial:        isPartial,
		gradingItemIds:   gradingItemIds,
		project:          project,
		dynamicFields:    dynamicFields,
		createdAt:        createdAt,
	}
}

func (e *Event) Topic() string           { return "judge_project_submitted" }
func (e *Event) ID() string              { return e.id }
func (e *Event) UserID() int32           { return e.userID }
func (e *Event) ContestProjectID() int32 { return e.contestProjectID }
func (e *Event) IsPartial() bool         { return e.isPartial }
func (e *Event) GradingItemIds() []int32 { return e.gradingItemIds }
func (e *Event) Project() *bytes.Buffer  { return e.project }
func (e *Event) DynamicFields() []*submissiondynamicfield.SubmissionDynamicField {
	return e.dynamicFields
}
func (e *Event) CreatedAt() time.Time { return e.createdAt }
