package scoringstrategy

var (
	DecreaseScoreStrategy = NewScoringStrategy("decrease", func(currentScore uint32, score uint32) uint32 {
		return currentScore - score
	})
)
