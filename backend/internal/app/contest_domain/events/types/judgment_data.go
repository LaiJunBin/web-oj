package types

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

type JudgmentData struct {
	SubmissionID string
	IsRejudge    bool
	Status       string
	Score        float64
	Reason       *string
	StaffNote    *string
	GradingItems []GradingItemData
}
