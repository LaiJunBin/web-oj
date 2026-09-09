package scoringstrategy

var (
	IncreaseScoreStrategy = NewScoringStrategy("increase", func(currentScore uint32, score uint32) uint32 {
		return currentScore + score
	})
)
