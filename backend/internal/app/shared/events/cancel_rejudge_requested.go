package sharedevent

import (
	sharedtypes "judge-backend/internal/app/shared/types"

	"github.com/google/uuid"
)

type CancelRejudgeRequested struct {
	id        string
	rejudgeID string
}

func NewCancelRejudgeRequested(rejudgeID string) sharedtypes.Event {
	return &CancelRejudgeRequested{
		id:        uuid.New().String(),
		rejudgeID: rejudgeID,
	}
}

func (e *CancelRejudgeRequested) Topic() string     { return "shared:cancel_rejudge_requested" }
func (e *CancelRejudgeRequested) ID() string        { return e.id }
func (e *CancelRejudgeRequested) RejudgeID() string { return e.rejudgeID }
