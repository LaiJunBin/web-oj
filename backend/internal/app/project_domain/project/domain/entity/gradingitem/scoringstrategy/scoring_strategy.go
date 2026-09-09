package scoringstrategy

import "errors"

var (
	Strategies = map[string]ScoringStrategy{
		IncreaseScoreStrategy.String(): IncreaseScoreStrategy,
		DecreaseScoreStrategy.String(): DecreaseScoreStrategy,
	}

	ErrorInvalidScoringStrategy = errors.New("invalid scoring strategy")
)

type ScoringStrategy interface {
	String() string
	Equals(other ScoringStrategy) bool
	CalculateScore(currentScore uint32, score uint32) uint32
}

type scoringStrategy struct {
	value     string
	calculate func(currentScore uint32, score uint32) uint32
}

func NewScoringStrategy(value string, calculate func(currentScore uint32, score uint32) uint32) ScoringStrategy {
	return &scoringStrategy{value: value, calculate: calculate}
}

func (s *scoringStrategy) String() string {
	return s.value
}

func (s *scoringStrategy) Equals(other ScoringStrategy) bool {
	return s.String() == other.String()
}

func (s *scoringStrategy) CalculateScore(currentScore uint32, score uint32) uint32 {
	return s.calculate(currentScore, score)
}
