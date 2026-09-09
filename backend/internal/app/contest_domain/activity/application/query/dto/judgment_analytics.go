package dto

type JudgmentStatusCount struct {
	Status string
	Count  int32
}

type JudgmentAnalytics struct {
	TotalSubmissions     int32
	TotalJudgments       int32
	StatusCounts         []*JudgmentStatusCount
	AvgJudgingDurationMs *float64
	AvgQueueWaitMs       *float64
}
