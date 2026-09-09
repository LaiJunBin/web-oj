package sharedevent

import sharedtypes "judge-backend/internal/app/shared/types"

type Judged struct {
	id               string
	judgmentID       string
	contestProjectID int32
}

func NewJudged(judgmentID string, contestProjectID int32) sharedtypes.Event {
	return &Judged{
		id:               judgmentID,
		judgmentID:       judgmentID,
		contestProjectID: contestProjectID,
	}
}

func (e *Judged) Topic() string           { return "shared:judged" }
func (e *Judged) ID() string              { return e.id }
func (e *Judged) JudgmentID() string      { return e.judgmentID }
func (e *Judged) ContestProjectID() int32 { return e.contestProjectID }
