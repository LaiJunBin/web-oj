package query

type PriorityStrategyQuery interface {
	GetStrategyNameByContestProjectID(contestProjectID int32) (string, error)
}
