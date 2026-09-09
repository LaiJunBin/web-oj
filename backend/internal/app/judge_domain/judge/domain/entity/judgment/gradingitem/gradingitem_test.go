package gradingitem_test

import (
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/gradingitem"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/gradingitem/scoringstrategy"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/status"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/testcase"
	"testing"
)

func newTestcase(id int32, score float64) *testcase.Testcase {
	return testcase.NewTestcase(&testcase.TestcaseConfig{
		ID:      id,
		Command: "echo test",
		Score:   score,
	})
}

func newPassedTestcase(id int32, score float64) *testcase.Testcase {
	tc := newTestcase(id, score)
	tc.SetStatus(status.PASSED)
	tc.SetParsedScore(score)
	return tc
}

func newFailedTestcase(id int32, score float64) *testcase.Testcase {
	tc := newTestcase(id, score)
	tc.SetStatus(status.FAILED)
	tc.SetParsedScore(score)
	return tc
}

func newRunningTestcase(id int32, score float64) *testcase.Testcase {
	tc := newTestcase(id, score)
	tc.SetStatus(status.RUNNING)
	return tc
}

func TestGradingItem(t *testing.T) {
	t.Run("NewGradingItem", func(t *testing.T) {
		t.Run("should store config values", func(t *testing.T) {
			tc := newTestcase(1, 10)
			gi := gradingitem.NewGradingItem(&gradingitem.GradingItemConfig{
				ID:              5,
				Score:           100,
				ScoringStrategy: scoringstrategy.IncreaseScoreStrategy,
				DependTestcases: []int32{1, 2},
				Testcases:       []*testcase.Testcase{tc},
			})

			if gi.ID() != 5 {
				t.Errorf("expected ID to be 5, got %d", gi.ID())
			}
			if gi.Score() != 100 {
				t.Errorf("expected score to be 100, got %f", gi.Score())
			}
			if len(gi.Testcases()) != 1 {
				t.Errorf("expected 1 testcase, got %d", len(gi.Testcases()))
			}
		})
	})

	t.Run("Status", func(t *testing.T) {
		t.Run("should return PASSED when no testcases", func(t *testing.T) {
			gi := gradingitem.NewGradingItem(&gradingitem.GradingItemConfig{
				ID:              1,
				Score:           100,
				ScoringStrategy: scoringstrategy.IncreaseScoreStrategy,
				Testcases:       []*testcase.Testcase{},
			})

			if !gi.Status().IsPassed() {
				t.Errorf("expected status to be Passed with no testcases, got %s", gi.Status().String())
			}
		})

		t.Run("should return PENDING when all testcases are pending", func(t *testing.T) {
			tc1 := newTestcase(1, 10) // default PENDING
			tc2 := newTestcase(2, 20) // default PENDING
			gi := gradingitem.NewGradingItem(&gradingitem.GradingItemConfig{
				ID:              1,
				Score:           100,
				ScoringStrategy: scoringstrategy.IncreaseScoreStrategy,
				Testcases:       []*testcase.Testcase{tc1, tc2},
			})

			if !gi.Status().IsPending() {
				t.Errorf("expected status to be Pending, got %s", gi.Status().String())
			}
		})

		t.Run("should return RUNNING when some testcases are running", func(t *testing.T) {
			tc1 := newPassedTestcase(1, 10)
			tc2 := newRunningTestcase(2, 20)
			gi := gradingitem.NewGradingItem(&gradingitem.GradingItemConfig{
				ID:              1,
				Score:           100,
				ScoringStrategy: scoringstrategy.IncreaseScoreStrategy,
				Testcases:       []*testcase.Testcase{tc1, tc2},
			})

			if !gi.Status().IsRunning() {
				t.Errorf("expected status to be Running, got %s", gi.Status().String())
			}
		})

		t.Run("should return RUNNING when some pending and some passed", func(t *testing.T) {
			tc1 := newPassedTestcase(1, 10)
			tc2 := newTestcase(2, 20) // PENDING
			gi := gradingitem.NewGradingItem(&gradingitem.GradingItemConfig{
				ID:              1,
				Score:           100,
				ScoringStrategy: scoringstrategy.IncreaseScoreStrategy,
				Testcases:       []*testcase.Testcase{tc1, tc2},
			})

			if !gi.Status().IsRunning() {
				t.Errorf("expected status to be Running when mix of pending and finished, got %s", gi.Status().String())
			}
		})

		t.Run("should return FAILED when any testcase failed", func(t *testing.T) {
			tc1 := newPassedTestcase(1, 10)
			tc2 := newFailedTestcase(2, 20)
			gi := gradingitem.NewGradingItem(&gradingitem.GradingItemConfig{
				ID:              1,
				Score:           100,
				ScoringStrategy: scoringstrategy.IncreaseScoreStrategy,
				Testcases:       []*testcase.Testcase{tc1, tc2},
			})

			if !gi.Status().IsFailed() {
				t.Errorf("expected status to be Failed, got %s", gi.Status().String())
			}
		})

		t.Run("should return FAILED when testcase has error status", func(t *testing.T) {
			tc1 := newTestcase(1, 10)
			tc1.SetStatus(status.ERROR)
			gi := gradingitem.NewGradingItem(&gradingitem.GradingItemConfig{
				ID:              1,
				Score:           100,
				ScoringStrategy: scoringstrategy.IncreaseScoreStrategy,
				Testcases:       []*testcase.Testcase{tc1},
			})

			// ERROR is !IsPassed() so it counts as failed
			if !gi.Status().IsFailed() {
				t.Errorf("expected status to be Failed for error testcase, got %s", gi.Status().String())
			}
		})

		t.Run("should return PASSED when all testcases passed", func(t *testing.T) {
			tc1 := newPassedTestcase(1, 10)
			tc2 := newPassedTestcase(2, 20)
			gi := gradingitem.NewGradingItem(&gradingitem.GradingItemConfig{
				ID:              1,
				Score:           100,
				ScoringStrategy: scoringstrategy.IncreaseScoreStrategy,
				Testcases:       []*testcase.Testcase{tc1, tc2},
			})

			if !gi.Status().IsPassed() {
				t.Errorf("expected status to be Passed, got %s", gi.Status().String())
			}
		})
	})

	t.Run("CalculateScore", func(t *testing.T) {
		t.Run("with increase strategy", func(t *testing.T) {
			t.Run("should sum passed testcase scores", func(t *testing.T) {
				tc1 := newPassedTestcase(1, 30)
				tc2 := newPassedTestcase(2, 40)
				gi := gradingitem.NewGradingItem(&gradingitem.GradingItemConfig{
					ID:              1,
					Score:           100,
					ScoringStrategy: scoringstrategy.IncreaseScoreStrategy,
					Testcases:       []*testcase.Testcase{tc1, tc2},
				})

				if gi.CalculateScore() != 70 {
					t.Errorf("expected score to be 70, got %f", gi.CalculateScore())
				}
			})

			t.Run("should cap at maxScore", func(t *testing.T) {
				tc1 := newPassedTestcase(1, 60)
				tc2 := newPassedTestcase(2, 60)
				gi := gradingitem.NewGradingItem(&gradingitem.GradingItemConfig{
					ID:              1,
					Score:           100,
					ScoringStrategy: scoringstrategy.IncreaseScoreStrategy,
					Testcases:       []*testcase.Testcase{tc1, tc2},
				})

				if gi.CalculateScore() != 100 {
					t.Errorf("expected score to be capped at 100, got %f", gi.CalculateScore())
				}
			})

			t.Run("should return 0 for no passed testcases", func(t *testing.T) {
				tc1 := newFailedTestcase(1, 50)
				gi := gradingitem.NewGradingItem(&gradingitem.GradingItemConfig{
					ID:              1,
					Score:           100,
					ScoringStrategy: scoringstrategy.IncreaseScoreStrategy,
					Testcases:       []*testcase.Testcase{tc1},
				})

				if gi.CalculateScore() != 0 {
					t.Errorf("expected score to be 0, got %f", gi.CalculateScore())
				}
			})
		})

		t.Run("with decrease strategy", func(t *testing.T) {
			t.Run("should return maxScore when all passed", func(t *testing.T) {
				tc1 := newPassedTestcase(1, 30)
				tc2 := newPassedTestcase(2, 40)
				gi := gradingitem.NewGradingItem(&gradingitem.GradingItemConfig{
					ID:              1,
					Score:           100,
					ScoringStrategy: scoringstrategy.DecreaseScoreStrategy,
					Testcases:       []*testcase.Testcase{tc1, tc2},
				})

				if gi.CalculateScore() != 100 {
					t.Errorf("expected score to be 100, got %f", gi.CalculateScore())
				}
			})

			t.Run("should subtract failed testcase scores", func(t *testing.T) {
				tc1 := newPassedTestcase(1, 30)
				tc2 := newFailedTestcase(2, 40)
				gi := gradingitem.NewGradingItem(&gradingitem.GradingItemConfig{
					ID:              1,
					Score:           100,
					ScoringStrategy: scoringstrategy.DecreaseScoreStrategy,
					Testcases:       []*testcase.Testcase{tc1, tc2},
				})

				if gi.CalculateScore() != 60 {
					t.Errorf("expected score to be 60, got %f", gi.CalculateScore())
				}
			})

			t.Run("should return 0 when pending testcases exist", func(t *testing.T) {
				tc1 := newPassedTestcase(1, 30)
				tc2 := newTestcase(2, 40) // PENDING
				gi := gradingitem.NewGradingItem(&gradingitem.GradingItemConfig{
					ID:              1,
					Score:           100,
					ScoringStrategy: scoringstrategy.DecreaseScoreStrategy,
					Testcases:       []*testcase.Testcase{tc1, tc2},
				})

				if gi.CalculateScore() != 0 {
					t.Errorf("expected score to be 0 when pending testcases exist, got %f", gi.CalculateScore())
				}
			})

			t.Run("should return 0 when cancelled testcases exist", func(t *testing.T) {
				tc1 := newPassedTestcase(1, 30)
				tc2 := newTestcase(2, 40)
				tc2.SetStatus(status.CANCELLED)
				gi := gradingitem.NewGradingItem(&gradingitem.GradingItemConfig{
					ID:              1,
					Score:           100,
					ScoringStrategy: scoringstrategy.DecreaseScoreStrategy,
					Testcases:       []*testcase.Testcase{tc1, tc2},
				})

				if gi.CalculateScore() != 0 {
					t.Errorf("expected score to be 0 when cancelled testcases exist, got %f", gi.CalculateScore())
				}
			})

			t.Run("should floor at 0 when deductions exceed maxScore", func(t *testing.T) {
				tc1 := newFailedTestcase(1, 60)
				tc2 := newFailedTestcase(2, 60)
				gi := gradingitem.NewGradingItem(&gradingitem.GradingItemConfig{
					ID:              1,
					Score:           100,
					ScoringStrategy: scoringstrategy.DecreaseScoreStrategy,
					Testcases:       []*testcase.Testcase{tc1, tc2},
				})

				if gi.CalculateScore() != 0 {
					t.Errorf("expected score to be floored at 0, got %f", gi.CalculateScore())
				}
			})
		})
	})

	t.Run("DependencyMet", func(t *testing.T) {
		t.Run("should return true when no dependencies", func(t *testing.T) {
			gi := gradingitem.NewGradingItem(&gradingitem.GradingItemConfig{
				ID:              1,
				Score:           100,
				ScoringStrategy: scoringstrategy.IncreaseScoreStrategy,
			})

			if !gi.DependencyMet(map[int32]struct{}{}) {
				t.Errorf("expected DependencyMet to be true with no dependencies")
			}
		})

		t.Run("should return true when all dependencies met", func(t *testing.T) {
			gi := gradingitem.NewGradingItem(&gradingitem.GradingItemConfig{
				ID:              3,
				Score:           100,
				ScoringStrategy: scoringstrategy.IncreaseScoreStrategy,
				DependTestcases: []int32{1, 2},
			})

			passed := map[int32]struct{}{1: {}, 2: {}, 5: {}}
			if !gi.DependencyMet(passed) {
				t.Errorf("expected DependencyMet to be true")
			}
		})

		t.Run("should return false when dependencies not met", func(t *testing.T) {
			gi := gradingitem.NewGradingItem(&gradingitem.GradingItemConfig{
				ID:              3,
				Score:           100,
				ScoringStrategy: scoringstrategy.IncreaseScoreStrategy,
				DependTestcases: []int32{1, 2},
			})

			passed := map[int32]struct{}{1: {}}
			if gi.DependencyMet(passed) {
				t.Errorf("expected DependencyMet to be false when dep 2 is missing")
			}
		})
	})

	t.Run("SetScore", func(t *testing.T) {
		gi := gradingitem.NewGradingItem(&gradingitem.GradingItemConfig{
			ID:              1,
			Score:           100,
			ScoringStrategy: scoringstrategy.IncreaseScoreStrategy,
		})

		gi.SetScore(50)
		if gi.Score() != 50 {
			t.Errorf("expected score to be 50, got %f", gi.Score())
		}
	})
}
