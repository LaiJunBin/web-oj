package sharedevent

import (
	sharedtypes "judge-backend/internal/app/shared/types"

	"github.com/google/uuid"
)

type CancelJudgeRequested struct {
	id         string
	judgmentID string
	reason     string
}

func NewCancelJudgeRequested(judgmentID, reason string) sharedtypes.Event {
	return &CancelJudgeRequested{
		id:         uuid.New().String(),
		judgmentID: judgmentID,
		reason:     reason,
	}
}

func (e *CancelJudgeRequested) Topic() string      { return "shared:cancel_judge_requested" }
func (e *CancelJudgeRequested) ID() string         { return e.id }
func (e *CancelJudgeRequested) JudgmentID() string { return e.judgmentID }
func (e *CancelJudgeRequested) Reason() string     { return e.reason }
