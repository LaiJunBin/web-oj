package sharedevent

import (
	"bytes"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/shared/types/submissiondynamicfield"
	"time"

	"github.com/google/uuid"
)

type ProjectSubmitted struct {
	id               string
	userID           int32
	contestProjectID int32
	isPartial        bool
	gradingItemIds   []int32
	project          *bytes.Buffer
	dynamicFields    []*submissiondynamicfield.SubmissionDynamicField
	createdAt        time.Time
}

func NewProjectSubmitted(
	userID int32,
	contestProjectID int32,
	isPartial bool,
	gradingItemIds []int32,
	project *bytes.Buffer,
	dynamicFields []*submissiondynamicfield.SubmissionDynamicField,
) sharedtypes.Event {
	return &ProjectSubmitted{
		id:               uuid.New().String(),
		userID:           userID,
		contestProjectID: contestProjectID,
		isPartial:        isPartial,
		gradingItemIds:   gradingItemIds,
		project:          project,
		dynamicFields:    dynamicFields,
		createdAt:        time.Now(),
	}
}

func (e *ProjectSubmitted) Topic() string           { return "shared:submit_project" }
func (e *ProjectSubmitted) ID() string              { return e.id }
func (e *ProjectSubmitted) UserID() int32           { return e.userID }
func (e *ProjectSubmitted) ContestProjectID() int32 { return e.contestProjectID }
func (e *ProjectSubmitted) IsPartial() bool         { return e.isPartial }
func (e *ProjectSubmitted) GradingItemIds() []int32 { return e.gradingItemIds }
func (e *ProjectSubmitted) Project() *bytes.Buffer  { return e.project }
func (e *ProjectSubmitted) DynamicFields() []*submissiondynamicfield.SubmissionDynamicField {
	return e.dynamicFields
}
func (e *ProjectSubmitted) CreatedAt() time.Time { return e.createdAt }
