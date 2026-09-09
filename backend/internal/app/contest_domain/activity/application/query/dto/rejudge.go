package dto

import "time"

type RejudgeInfo struct {
	ID                string
	ContestProjectID  int32
	Reason            string
	CreatedAt         time.Time
	Status            string
	EffectedJudgments int32

	// CompletedJudgments is how many of EffectedJudgments have reached a
	// terminal state. Together the two are a progress bar; on their own,
	// EffectedJudgments says nothing about how far along the rejudge is.
	CompletedJudgments int32
}

type JudgmentOutcomeChange struct {
	Status     string
	Score      float64
	TotalScore float64
}

type SubmissionComparison struct {
	ID string

	// Before is nil when the submission had no adopted judgment before the
	// rejudge ran — a pointer rather than a value so "no baseline" cannot be
	// mistaken for a genuine zero score.
	Before *JudgmentOutcomeChange
	After  JudgmentOutcomeChange
}

type Rejudge struct {
	ID          string
	Submissions []*SubmissionComparison
	CreatedAt   time.Time
	Reason      string
	Status      string
}

// TestcaseOutcome is one testcase as a judgment recorded it. Score is what the
// testcase earned, TotalScore its configured max, and ParsedScore the raw parser
// output the earned score derives from.
type TestcaseOutcome struct {
	ID          int32
	Name        string
	Status      string
	Score       float64
	TotalScore  float64
	ParsedScore float64
}

type GradingItemOutcome struct {
	ID              int32
	Label           string
	Name            string
	Status          string
	Score           float64
	TotalScore      float64
	ScoringStrategy string
	Testcases       []*TestcaseOutcome
}

type JudgmentOutcomeDetail struct {
	JudgmentID   string
	Status       string
	Score        float64
	TotalScore   float64
	GradingItems []*GradingItemOutcome
}

// SubmissionComparisonDetail is SubmissionComparison at grading-item and
// testcase resolution. Before is nil when the submission had no adopted judgment
// before the rejudge ran.
//
// Grading items and testcases are matched between the two sides by ID, never by
// position: a rejudge is usually triggered *because* the project's grading
// configuration changed, so an id can legitimately exist on only one side.
type SubmissionComparisonDetail struct {
	ID     string
	Before *JudgmentOutcomeDetail
	After  *JudgmentOutcomeDetail
}
