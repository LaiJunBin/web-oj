package judgment

import (
	"fmt"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/gradingitem"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/judgestatus"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/judgetype"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/lifecycle"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/project"
	"time"
)

type Judgment struct {
	id                string
	contestProjectID  int32
	project           *project.Project
	submissionID      string
	rejudgeID         *string
	gradingItems      []*gradingitem.GradingItem
	lifecycle         *lifecycle.JudgmentLifecycle
	status            judgestatus.JudgeStatus
	judgeType         judgetype.JudgeType
	score             float64
	errorMessage      *string
	reason            *string
	staffNote         *string
	createdAt         time.Time
	judgingStartedAt  *time.Time
	judgingFinishedAt *time.Time

	// maxJudgeDurationMinutes is the contest project's override of the
	// worker's judgment duration budget. Transient: resolved from the DB by
	// the enqueue handler each time the judgment is (re)queued — including
	// after a restart — and never persisted with the judgment document.
	maxJudgeDurationMinutes *int32

	sortingFields map[string]interface{}
}

type JudgmentConfig struct {
	ContestProjectID  int32
	Project           *project.Project
	SubmissionID      string
	GradingItems      []*gradingitem.GradingItem
	Lifecycle         *lifecycle.JudgmentLifecycle
	Status            *judgestatus.JudgeStatus
	JudgeType         judgetype.JudgeType
	Reason            *string
	CreatedAt         time.Time
	JudgingStartedAt  *time.Time
	JudgingFinishedAt *time.Time
}

func New(config *JudgmentConfig) *Judgment {
	status := *judgestatus.QUEUED
	if config.Status != nil {
		status = *config.Status
	}

	createdAt := time.Now()
	if !config.CreatedAt.IsZero() {
		createdAt = config.CreatedAt
	}

	score := 0.0
	for _, gradingItem := range config.GradingItems {
		score += gradingItem.Score()
	}

	return &Judgment{
		contestProjectID:  config.ContestProjectID,
		project:           config.Project,
		submissionID:      config.SubmissionID,
		gradingItems:      config.GradingItems,
		lifecycle:         config.Lifecycle,
		status:            status,
		score:             score,
		judgeType:         config.JudgeType,
		reason:            config.Reason,
		createdAt:         createdAt,
		judgingStartedAt:  config.JudgingStartedAt,
		judgingFinishedAt: config.JudgingFinishedAt,
	}
}

func (j *Judgment) ID() string {
	return j.id
}

func (j *Judgment) ContestProjectID() int32 {
	return j.contestProjectID
}

func (j *Judgment) MaxJudgeDurationMinutes() *int32 {
	return j.maxJudgeDurationMinutes
}

func (j *Judgment) SetMaxJudgeDurationMinutes(minutes *int32) {
	j.maxJudgeDurationMinutes = minutes
}

func (j *Judgment) Project() *project.Project {
	return j.project
}

func (j *Judgment) SubmissionID() string {
	return j.submissionID
}

func (j *Judgment) RejudgeID() *string {
	return j.rejudgeID
}

func (j *Judgment) GradingItems() []*gradingitem.GradingItem {
	return j.gradingItems
}

func (j *Judgment) Lifecycle() *lifecycle.JudgmentLifecycle {
	return j.lifecycle
}

func (j *Judgment) Status() judgestatus.JudgeStatus {
	return j.status
}

func (j *Judgment) Score() float64 {
	return j.score
}

func (j *Judgment) CalculateScore() float64 {
	score := 0.0

	for _, gradingItem := range j.gradingItems {
		score += gradingItem.CalculateScore()
	}

	return score
}

func (j *Judgment) JudgeType() judgetype.JudgeType {
	return j.judgeType
}

func (j *Judgment) ErrorMessage() *string {
	return j.errorMessage
}

func (j *Judgment) Reason() *string {
	return j.reason
}

// StaffNote is a judge's note to the contestant, written after looking into
// the judgment; it sits beside Reason and never replaces it. Nil = none.
func (j *Judgment) StaffNote() *string {
	return j.staffNote
}

func (j *Judgment) CreatedAt() time.Time {
	return j.createdAt
}

func (j *Judgment) JudgingStartedAt() *time.Time {
	return j.judgingStartedAt
}

func (j *Judgment) JudgingFinishedAt() *time.Time {
	return j.judgingFinishedAt
}

func (j *Judgment) SortingFields() map[string]interface{} {
	return j.sortingFields
}

func (j *Judgment) SetID(id string) {
	j.id = id
}

// --- Guarded state transitions ---
//
// These are the only sanctioned ways to move a judgment through its lifecycle:
//
//	QUEUED → JUDGING → JUDGED / REJUDGED
//	QUEUED / JUDGING → CANCELLED
//	QUEUED / JUDGING → ERROR
//
// SetStatus remains available for repository hydration only.

// ErrInvalidTransition is returned when a lifecycle method is called on a
// judgment whose current status does not allow that transition.
type ErrInvalidTransition struct {
	From string
	To   string
}

func (e *ErrInvalidTransition) Error() string {
	return fmt.Sprintf("invalid judgment status transition from %s to %s", e.From, e.To)
}

// Start moves a queued judgment into judging and stamps judgingStartedAt.
func (j *Judgment) Start() error {
	if !j.status.IsQueued() {
		return &ErrInvalidTransition{From: j.status.String(), To: judgestatus.JUDGING.String()}
	}
	j.status = *judgestatus.JUDGING
	now := time.Now()
	j.judgingStartedAt = &now
	return nil
}

// Finish moves a judging judgment into its terminal success status
// (JUDGED, or REJUDGED for rejudge-type judgments) and stamps judgingFinishedAt.
func (j *Judgment) Finish() error {
	target := judgestatus.JUDGED
	if j.judgeType.IsRejudge() {
		target = judgestatus.REJUDGED
	}
	if !j.status.IsJudging() {
		return &ErrInvalidTransition{From: j.status.String(), To: target.String()}
	}
	j.status = *target
	now := time.Now()
	j.judgingFinishedAt = &now
	return nil
}

// Cancel moves a queued or judging judgment into CANCELLED. A non-empty
// reason overwrites the current one; an empty reason keeps it untouched.
func (j *Judgment) Cancel(reason string) error {
	if !j.status.IsQueued() && !j.status.IsJudging() {
		return &ErrInvalidTransition{From: j.status.String(), To: judgestatus.CANCELLED.String()}
	}
	j.status = *judgestatus.CANCELLED
	if reason != "" {
		j.reason = &reason
	}
	now := time.Now()
	j.judgingFinishedAt = &now
	return nil
}

// Fail moves a queued or judging judgment into ERROR and stamps judgingFinishedAt.
func (j *Judgment) Fail() error {
	if !j.status.IsQueued() && !j.status.IsJudging() {
		return &ErrInvalidTransition{From: j.status.String(), To: judgestatus.ERROR.String()}
	}
	j.status = *judgestatus.ERROR
	now := time.Now()
	j.judgingFinishedAt = &now
	return nil
}

func (j *Judgment) SetStatus(status *judgestatus.JudgeStatus) {
	j.status = *status
}

func (j *Judgment) SetRejudgeID(rejudgeID string) {
	j.rejudgeID = &rejudgeID
}

func (j *Judgment) SetScore(score float64) {
	j.score = score
}

func (j *Judgment) SetErrorMessage(errorMessage string) {
	j.errorMessage = &errorMessage
}

// SetStaffNote takes a pointer on purpose: nil CLEARS the note. SetReason's
// value signature cannot express clearing, and a staff note must be
// retractable.
func (j *Judgment) SetStaffNote(note *string) {
	j.staffNote = note
}

func (j *Judgment) SetReason(reason string) {
	j.reason = &reason
}

func (j *Judgment) SetJudgingStartedAt(startedAt time.Time) {
	j.judgingStartedAt = &startedAt
}

func (j *Judgment) SetJudgingFinishedAt(finishedAt time.Time) {
	j.judgingFinishedAt = &finishedAt
}

func (j *Judgment) SetSortingFields(s map[string]interface{}) {
	j.sortingFields = s
}
