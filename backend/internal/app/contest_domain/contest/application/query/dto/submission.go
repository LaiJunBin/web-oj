package dto

import "time"

type Submission struct {
	ID         string
	Status     string
	IsPartial  bool
	Score      float64
	TotalScore float64
	CreatedAt  time.Time
	// Reason is the contestant-facing explanation a judgment recorded when it
	// ended in Error (or the reason given for a cancellation); nil otherwise.
	// Every writer keeps it neutral — no commands, paths or internal detail —
	// because the user API now serialises it. Internal detail lives in
	// errorMessage, which stays on the admin API only.
	Reason *string
	// StaffNote is the operator's note to the contestant, shown next to
	// Reason on the list as well as the detail — the list is where a
	// contestant first looks, and a note they only find by opening the
	// detail is a note they miss. nil when there is none.
	StaffNote *string
}

type SubmissionDetail struct {
	Submission
	GradingItems []SubmissionGradingItem
	// SampleTestcaseIDs are the ids of the testcases marked is_sample in THIS
	// submission's problem — the only testcases whose run output a contestant
	// may see. The REST detail already filters its testcases with it; the
	// live stream uses it to apply the same rule to pushed frames. Internal:
	// the REST presenter maps fields explicitly and never serialises it.
	SampleTestcaseIDs []int32
	// ContestProjectID names which contest project the submission belongs
	// to, so the stream can re-resolve the sample set while it stays open
	// (testcase visibility CAN change mid-contest). Internal, like the field
	// above: never serialised by the REST presenter.
	ContestProjectID int32
}

type SubmissionGradingItem struct {
	ID        int32
	Score     float64
	Status    string
	Testcases []SubmissionTestcase
}

type SubmissionTestcase struct {
	ID       int32
	ExecTime int64
	ExitCode int32
	Stdout   string
	Stderr   string
	Status   string
}
