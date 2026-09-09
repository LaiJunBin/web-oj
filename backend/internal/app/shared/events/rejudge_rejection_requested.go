package sharedevent

import (
	sharedtypes "judge-backend/internal/app/shared/types"

	"github.com/google/uuid"
)

type RejudgeRejectionRequested struct {
	id        string
	rejudgeID string
}

func NewRejudgeRejectionRequested(rejudgeID string) sharedtypes.Event {
	return &RejudgeRejectionRequested{
		id:        uuid.New().String(),
		rejudgeID: rejudgeID,
	}
}

func (e *RejudgeRejectionRequested) Topic() string     { return "shared:rejudge_reject_requested" }
func (e *RejudgeRejectionRequested) ID() string        { return e.id }
func (e *RejudgeRejectionRequested) RejudgeID() string { return e.rejudgeID }
