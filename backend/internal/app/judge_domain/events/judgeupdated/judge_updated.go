package judgeupdated

import (
	"fmt"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment"

	"github.com/google/uuid"
)

type TestcaseData struct {
	ID        int32
	Status    string
	ExecTime  int64
	ExitCode  int
	Stdout    string
	Stderr    string
	HasResult bool
}

type GradingItemData struct {
	ID        int32
	Score     float64
	Status    string
	Testcases []TestcaseData
}

type JudgeUpdatedData struct {
	JudgmentID       string
	RejudgeID        *string
	ContestProjectID int32
	SubmissionID     string
	IsRejudge        bool
	Status           string
	StatusIsError    bool
	// Reason is the judgment's contestant-facing explanation (Error/Cancelled),
	// carried so the live stream can show it without a reload.
	Reason            *string
	// StaffNote is the operator-written note shown to the contestant next to
	// Reason. nil means "no note" — and, on a frame that follows a clear, it
	// means "the note is gone", so consumers must not keep a stale value.
	StaffNote         *string
	StatusIsCancelled bool
	StatusIsRejudged  bool
	Score             float64
	GradingItems      []GradingItemData
}

type JudgeUpdatedEvent struct {
	id   string
	data JudgeUpdatedData
}

func (e *JudgeUpdatedEvent) Topic() string {
	return "judge_updated"
}

func NewJudgeUpdatedEvent(j *judgment.Judgment) *JudgeUpdatedEvent {
	gradingItems := make([]GradingItemData, 0, len(j.GradingItems()))
	for _, gi := range j.GradingItems() {
		testcases := make([]TestcaseData, 0, len(gi.Testcases()))
		for _, tc := range gi.Testcases() {
			td := TestcaseData{
				ID:     tc.ID(),
				Status: tc.Status().String(),
			}
			if r := tc.Result(); r != nil {
				td.HasResult = true
				td.ExecTime = r.ExecTime()
				td.ExitCode = r.ExitCode()
				td.Stdout = r.Stdout()
				td.Stderr = r.Stderr()
			}
			testcases = append(testcases, td)
		}
		gradingItems = append(gradingItems, GradingItemData{
			ID:        gi.ID(),
			Score:     gi.CalculateScore(),
			Status:    gi.Status().String(),
			Testcases: testcases,
		})
	}

	return &JudgeUpdatedEvent{
		// The ID must be unique per publish: a judgment emits many updates
		// during its life, and any claim-based dedupe downstream would treat
		// a per-judgment ID as "already processed", swallowing every update
		// after the first — which is exactly how the rejudge status flip was
		// once lost. Today's subscribers deliberately do not claim this
		// event, but the unique ID keeps that mistake impossible to reintroduce.
		id: fmt.Sprintf("%s-%s-%s", "judge_updated", j.ID(), uuid.New().String()),
		data: JudgeUpdatedData{
			JudgmentID:        j.ID(),
			RejudgeID:         j.RejudgeID(),
			ContestProjectID:  j.ContestProjectID(),
			SubmissionID:      j.SubmissionID(),
			IsRejudge:         j.JudgeType().IsRejudge(),
			Status:            j.Status().String(),
			StatusIsError:     j.Status().IsError(),
			Reason:            j.Reason(),
			StaffNote:         j.StaffNote(),
			StatusIsCancelled: j.Status().IsCancelled(),
			StatusIsRejudged:  j.Status().IsRejudged(),
			Score:             j.CalculateScore(),
			GradingItems:      gradingItems,
		},
	}
}

func (e *JudgeUpdatedEvent) ID() string {
	return e.id
}

func (e *JudgeUpdatedEvent) Data() JudgeUpdatedData {
	return e.data
}
