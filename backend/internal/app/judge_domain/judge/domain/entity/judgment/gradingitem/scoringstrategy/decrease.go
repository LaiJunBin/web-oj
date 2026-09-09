package scoringstrategy

import "judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/testcase"

var (
	DecreaseScoreStrategy = NewScoringStrategy(
		"decrease",
		func(testcases []*testcase.Testcase, maxScore float64) float64 {
			score := maxScore
			for _, testcase := range testcases {
				if testcase.Status().IsPending() || testcase.Status().IsCancelled() {
					score = 0
					break
				}

				if !testcase.Status().IsPassed() {
					score -= testcase.ParsedScore()
				}
			}

			if score < 0 {
				score = 0
			}

			return score
		},
		// Under "decrease" the item starts at its max score and every non-passing
		// testcase deducts whatever the parser charged against it. Per testcase
		// that reads as: keep your configured score when you pass, lose the
		// parser's deduction when you don't, and earn nothing when you never
		// produced a verdict.
		//
		// Two places this view diverges from CalculateScore, both worth surfacing
		// rather than papering over:
		//   - CalculateScore zeroes the WHOLE item when any testcase is pending or
		//     cancelled. A single testcase can't see its siblings, so the other
		//     testcases still report what they individually earned.
		//   - CalculateScore deducts from the item's max, not from the sum of the
		//     testcases' configured scores. When those two differ, the parts don't
		//     add back up to the item total.
		func(tc *testcase.Testcase) float64 {
			tcStatus := tc.Status()
			if tcStatus.IsPending() || tcStatus.IsRunning() || tcStatus.IsCancelled() {
				return 0
			}

			if tcStatus.IsPassed() {
				return tc.Score()
			}

			score := tc.Score() - tc.ParsedScore()
			if score < 0 {
				return 0
			}
			return score
		},
	)
)
