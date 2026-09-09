package query

// MaxJudgeDurationQuery resolves the per-contest-project cap on one
// judgment's total wall-clock. Nil means the contest project does not
// override the global judge config default.
type MaxJudgeDurationQuery interface {
	GetMaxJudgeDurationMinutesByContestProjectID(contestProjectID int32) (*int32, error)
}
