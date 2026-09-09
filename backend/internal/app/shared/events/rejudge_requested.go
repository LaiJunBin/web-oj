package sharedevent

import (
	sharedtypes "judge-backend/internal/app/shared/types"

	"github.com/google/uuid"
)

type RejudgeRequested struct {
	id               string
	contestProjectID int32
	submissionIDs    []string
	reason           string
}

// NewRejudgeRequested carries the contest project explicitly rather than
// leaving it to be derived from the submissions downstream: a whole-project
// rejudge over a project with no submissions still creates a rejudge, and there
// would be nothing to derive it from.
func NewRejudgeRequested(contestProjectID int32, submissionIDs []string, reason string) sharedtypes.Event {
	return &RejudgeRequested{
		id:               uuid.New().String(),
		contestProjectID: contestProjectID,
		submissionIDs:    submissionIDs,
		reason:           reason,
	}
}

func (e *RejudgeRequested) Topic() string           { return "shared:rejudge_requested" }
func (e *RejudgeRequested) ID() string              { return e.id }
func (e *RejudgeRequested) ContestProjectID() int32 { return e.contestProjectID }
func (e *RejudgeRequested) SubmissionIDs() []string { return e.submissionIDs }
func (e *RejudgeRequested) Reason() string          { return e.reason }
