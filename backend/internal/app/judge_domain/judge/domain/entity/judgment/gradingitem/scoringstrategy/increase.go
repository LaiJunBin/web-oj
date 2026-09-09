package scoringstrategy

import "judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/testcase"

var (
	IncreaseScoreStrategy = NewScoringStrategy(
		"increase",
		func(testcases []*testcase.Testcase, maxScore float64) float64 {
			score := 0.0
			for _, testcase := range testcases {
				if testcase.Status().IsPassed() {
					score += testcase.ParsedScore()
				}
			}

			if score > maxScore {
				score = maxScore
			}

			return score
		},
		// Under "increase" a testcase contributes exactly what the parser awarded
		// it, and only when it passed. Anything else — failed, errored, timed out,
		// or never run — contributes nothing.
		//
		// Deliberately NOT clamped to the testcase's configured score: the stdout
		// parser reads the score out of the testcase's own output, so it can
		// legitimately exceed the configured value, and clamping would hide a
		// misconfigured grading item. CalculateScore clamps the *sum* at the item's
		// max instead, so an overshooting set is exactly where the per-testcase
		// parts stop adding up to the item total.
		func(tc *testcase.Testcase) float64 {
			if !tc.Status().IsPassed() {
				return 0
			}
			return tc.ParsedScore()
		},
	)
)
