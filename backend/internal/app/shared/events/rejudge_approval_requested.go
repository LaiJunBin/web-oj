package sharedevent

import (
	sharedtypes "judge-backend/internal/app/shared/types"

	"github.com/google/uuid"
)

type RejudgeApprovalRequested struct {
	id        string
	rejudgeID string
}

func NewRejudgeApprovalRequested(rejudgeID string) sharedtypes.Event {
	return &RejudgeApprovalRequested{
		id:        uuid.New().String(),
		rejudgeID: rejudgeID,
	}
}

func (e *RejudgeApprovalRequested) Topic() string     { return "shared:rejudge_approval_requested" }
func (e *RejudgeApprovalRequested) ID() string        { return e.id }
func (e *RejudgeApprovalRequested) RejudgeID() string { return e.rejudgeID }
