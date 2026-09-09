package sharedevent

import (
	"fmt"
	sharedtypes "judge-backend/internal/app/shared/types"

	"github.com/google/uuid"
)

type JudgeUpdatedTestcaseData struct {
	ID        int32
	Status    string
	ExecTime  int64
	ExitCode  int
	Stdout    string
	Stderr    string
	HasResult bool
}

type JudgeUpdatedGradingItemData struct {
	ID        int32
	Score     float64
	Status    string
	Testcases []JudgeUpdatedTestcaseData
}

type JudgeUpdatedData struct {
	JudgmentID        string
	RejudgeID         *string
	ContestProjectID  int32
	SubmissionID      string
	IsRejudge         bool
	Status            string
	StatusIsError     bool
	Reason            *string
	StaffNote         *string
	StatusIsCancelled bool
	StatusIsRejudged  bool
	Score             float64
	GradingItems      []JudgeUpdatedGradingItemData
}

type JudgeUpdated struct {
	id   string
	data JudgeUpdatedData
}

func NewJudgeUpdated(data JudgeUpdatedData) sharedtypes.Event {
	return &JudgeUpdated{
		// The ID must be unique per publish: a judgment emits many updates
		// during its life. A per-judgment ID lets downstream claim-based
		// dedup (TryClaimEvent) swallow every update after the first — that
		// is how contest-side submission updates were silently dropped.
		id:   fmt.Sprintf("judge_updated-%s-%s", data.JudgmentID, uuid.New().String()),
		data: data,
	}
}

func (e *JudgeUpdated) Topic() string          { return "shared:judge_updated" }
func (e *JudgeUpdated) ID() string             { return e.id }
func (e *JudgeUpdated) Data() JudgeUpdatedData { return e.data }
