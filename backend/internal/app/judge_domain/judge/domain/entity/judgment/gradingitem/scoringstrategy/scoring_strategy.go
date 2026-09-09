package scoringstrategy

import (
	"errors"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/testcase"
)

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
	CalculateScore(testcases []*testcase.Testcase, maxScore float64) float64

	// TestcaseScore decomposes a grading item's score onto one testcase, so a
	// report can show what each testcase contributed rather than only the item
	// total. It is a per-testcase view of the same arithmetic CalculateScore
	// runs over the whole set, which means the parts only add back up to
	// CalculateScore when the testcases' configured scores sum to the item's
	// max score — see each strategy for where the two can diverge.
	TestcaseScore(tc *testcase.Testcase) float64
}

type scoringStrategy struct {
	value         string
	calculate     func(testcases []*testcase.Testcase, maxScore float64) float64
	testcaseScore func(tc *testcase.Testcase) float64
}

func NewScoringStrategy(
	value string,
	calculate func(testcases []*testcase.Testcase, maxScore float64) float64,
	testcaseScore func(tc *testcase.Testcase) float64,
) ScoringStrategy {
	return &scoringStrategy{value: value, calculate: calculate, testcaseScore: testcaseScore}
}

func (s *scoringStrategy) String() string {
	return s.value
}

func (s *scoringStrategy) Equals(other ScoringStrategy) bool {
	return s.String() == other.String()
}

func (s *scoringStrategy) CalculateScore(testcases []*testcase.Testcase, maxScore float64) float64 {
	return s.calculate(testcases, maxScore)
}

func (s *scoringStrategy) TestcaseScore(tc *testcase.Testcase) float64 {
	return s.testcaseScore(tc)
}
