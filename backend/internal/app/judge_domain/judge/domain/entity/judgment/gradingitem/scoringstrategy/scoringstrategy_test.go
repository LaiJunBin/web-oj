package scoringstrategy_test

import (
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/gradingitem/scoringstrategy"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/status"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/testcase"
	"testing"
)

func makeTestcase(id int32, configuredScore, parsedScore float64, s status.Status) *testcase.Testcase {
	tc := testcase.NewTestcase(&testcase.TestcaseConfig{
		ID:          id,
		Score:       configuredScore,
		ParsedScore: parsedScore,
	})
	tc.SetStatus(s)
	tc.SetParsedScore(parsedScore)
	return tc
}

// ── ExitCode + Increase ────────────────────────────────────────────────────────
// TC1 PASSED → +10 → 10
// TC2 FAILED → skip → 10
// TC3 PASSED → +10 → 20
// Final: 20 / 30
func TestIncrease_ExitCode(t *testing.T) {
	testcases := []*testcase.Testcase{
		makeTestcase(1, 10, 10, status.PASSED),
		makeTestcase(2, 10, 10, status.FAILED),
		makeTestcase(3, 10, 10, status.PASSED),
	}
	got := scoringstrategy.IncreaseScoreStrategy.CalculateScore(testcases, 30)
	if got != 20 {
		t.Errorf("expected 20, got %v", got)
	}
}

// ── ExitCode + Decrease ────────────────────────────────────────────────────────
// TC1 PASSED → skip → 30
// TC2 FAILED → -10 → 20
// TC3 PASSED → skip → 20
// Final: 20 / 30
func TestDecrease_ExitCode(t *testing.T) {
	testcases := []*testcase.Testcase{
		makeTestcase(1, 10, 10, status.PASSED),
		makeTestcase(2, 10, 10, status.FAILED),
		makeTestcase(3, 10, 10, status.PASSED),
	}
	got := scoringstrategy.DecreaseScoreStrategy.CalculateScore(testcases, 30)
	if got != 20 {
		t.Errorf("expected 20, got %v", got)
	}
}

// ── Stdout + Increase ──────────────────────────────────────────────────────────
// TC1 "pass 8"  → PASSED, parsedScore=8  → +8 → 8
// TC2 "fail 3"  → FAILED, parsedScore=3  → skip → 8
// TC3 "pass"    → PASSED, parsedScore=10 → +10 → 18
// Final: 18 / 30
func TestIncrease_Stdout(t *testing.T) {
	testcases := []*testcase.Testcase{
		makeTestcase(1, 10, 8, status.PASSED),
		makeTestcase(2, 10, 3, status.FAILED),
		makeTestcase(3, 10, 10, status.PASSED),
	}
	got := scoringstrategy.IncreaseScoreStrategy.CalculateScore(testcases, 30)
	if got != 18 {
		t.Errorf("expected 18, got %v", got)
	}
}

// ── Stdout + Decrease ──────────────────────────────────────────────────────────
// TC1 "pass 8"  → PASSED, parsedScore=8  → skip → 30
// TC2 "fail 3"  → FAILED, parsedScore=3  → -3 → 27
// TC3 "fail"    → FAILED, parsedScore=10 → -10 → 17
// Final: 17 / 30
func TestDecrease_Stdout(t *testing.T) {
	testcases := []*testcase.Testcase{
		makeTestcase(1, 10, 8, status.PASSED),
		makeTestcase(2, 10, 3, status.FAILED),
		makeTestcase(3, 10, 10, status.FAILED),
	}
	got := scoringstrategy.DecreaseScoreStrategy.CalculateScore(testcases, 30)
	if got != 17 {
		t.Errorf("expected 17, got %v", got)
	}
}

// ── Edge cases ─────────────────────────────────────────────────────────────────

func TestIncrease_AllPassed(t *testing.T) {
	testcases := []*testcase.Testcase{
		makeTestcase(1, 10, 10, status.PASSED),
		makeTestcase(2, 10, 10, status.PASSED),
		makeTestcase(3, 10, 10, status.PASSED),
	}
	got := scoringstrategy.IncreaseScoreStrategy.CalculateScore(testcases, 30)
	if got != 30 {
		t.Errorf("expected 30, got %v", got)
	}
}

func TestDecrease_AllPassed(t *testing.T) {
	testcases := []*testcase.Testcase{
		makeTestcase(1, 10, 10, status.PASSED),
		makeTestcase(2, 10, 10, status.PASSED),
		makeTestcase(3, 10, 10, status.PASSED),
	}
	got := scoringstrategy.DecreaseScoreStrategy.CalculateScore(testcases, 30)
	if got != 30 {
		t.Errorf("expected 30, got %v", got)
	}
}

func TestIncrease_Cancelled_ZeroScore(t *testing.T) {
	testcases := []*testcase.Testcase{
		makeTestcase(1, 10, 10, status.PASSED),
		makeTestcase(2, 10, 10, status.CANCELLED),
		makeTestcase(3, 10, 10, status.PASSED),
	}
	got := scoringstrategy.IncreaseScoreStrategy.CalculateScore(testcases, 30)
	if got != 20 {
		t.Errorf("expected 20 (cancelled does not add), got %v", got)
	}
}

func TestDecrease_Pending_ZeroScore(t *testing.T) {
	testcases := []*testcase.Testcase{
		makeTestcase(1, 10, 10, status.PASSED),
		makeTestcase(2, 10, 10, status.PENDING),
		makeTestcase(3, 10, 10, status.PASSED),
	}
	got := scoringstrategy.DecreaseScoreStrategy.CalculateScore(testcases, 30)
	if got != 0 {
		t.Errorf("expected 0 (pending breaks to zero), got %v", got)
	}
}

func TestDecrease_ScoreFloorZero(t *testing.T) {
	testcases := []*testcase.Testcase{
		makeTestcase(1, 10, 10, status.FAILED),
		makeTestcase(2, 10, 10, status.FAILED),
		makeTestcase(3, 10, 10, status.FAILED),
		makeTestcase(4, 10, 10, status.FAILED),
	}
	got := scoringstrategy.DecreaseScoreStrategy.CalculateScore(testcases, 30)
	if got != 0 {
		t.Errorf("expected 0 (floor), got %v", got)
	}
}

// ── TestcaseScore: per-testcase decomposition ──────────────────────────────────

func TestIncrease_TestcaseScore(t *testing.T) {
	cases := []struct {
		name     string
		testcase *testcase.Testcase
		want     float64
	}{
		{"passed awards the parsed score", makeTestcase(1, 10, 8, status.PASSED), 8},
		{"failed awards nothing", makeTestcase(1, 10, 3, status.FAILED), 0},
		{"errored awards nothing", makeTestcase(1, 10, 3, status.ERROR), 0},
		{"timed out awards nothing", makeTestcase(1, 10, 3, status.TIMEOUT), 0},
		{"pending awards nothing", makeTestcase(1, 10, 0, status.PENDING), 0},
		{"running awards nothing", makeTestcase(1, 10, 0, status.RUNNING), 0},
		{"cancelled awards nothing", makeTestcase(1, 10, 10, status.CANCELLED), 0},
		// The stdout parser reads the score out of the testcase's own output, so it
		// can exceed the configured score. That is reported as-is rather than
		// clamped, so a misconfigured item stays visible.
		{"parsed score above the configured max is not clamped", makeTestcase(1, 10, 99, status.PASSED), 99},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := scoringstrategy.IncreaseScoreStrategy.TestcaseScore(c.testcase)
			if got != c.want {
				t.Errorf("expected %v, got %v", c.want, got)
			}
		})
	}
}

func TestDecrease_TestcaseScore(t *testing.T) {
	cases := []struct {
		name     string
		testcase *testcase.Testcase
		want     float64
	}{
		{"passed keeps the configured score", makeTestcase(1, 10, 8, status.PASSED), 10},
		{"failed loses the parser's deduction", makeTestcase(1, 10, 3, status.FAILED), 7},
		{"errored loses the parser's deduction", makeTestcase(1, 10, 4, status.ERROR), 6},
		{"timed out loses the parser's deduction", makeTestcase(1, 10, 10, status.TIMEOUT), 0},
		{"pending awards nothing", makeTestcase(1, 10, 0, status.PENDING), 0},
		{"running awards nothing", makeTestcase(1, 10, 0, status.RUNNING), 0},
		{"cancelled awards nothing", makeTestcase(1, 10, 0, status.CANCELLED), 0},
		{"a deduction beyond the configured score floors at zero", makeTestcase(1, 10, 99, status.FAILED), 0},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := scoringstrategy.DecreaseScoreStrategy.TestcaseScore(c.testcase)
			if got != c.want {
				t.Errorf("expected %v, got %v", c.want, got)
			}
		})
	}
}

// The per-testcase decomposition is only useful if it adds back up to the item
// total, so pin that for the well-configured case both strategies are designed
// around: testcase scores summing to the item's max.
func TestTestcaseScore_SumsToCalculateScore(t *testing.T) {
	t.Run("increase", func(t *testing.T) {
		testcases := []*testcase.Testcase{
			makeTestcase(1, 10, 8, status.PASSED),
			makeTestcase(2, 10, 3, status.FAILED),
			makeTestcase(3, 10, 10, status.PASSED),
		}

		sum := 0.0
		for _, tc := range testcases {
			sum += scoringstrategy.IncreaseScoreStrategy.TestcaseScore(tc)
		}

		item := scoringstrategy.IncreaseScoreStrategy.CalculateScore(testcases, 30)
		if sum != item {
			t.Errorf("expected per-testcase scores to sum to the item score %v, got %v", item, sum)
		}
	})

	t.Run("decrease", func(t *testing.T) {
		testcases := []*testcase.Testcase{
			makeTestcase(1, 10, 8, status.PASSED),
			makeTestcase(2, 10, 3, status.FAILED),
			makeTestcase(3, 10, 10, status.FAILED),
		}

		sum := 0.0
		for _, tc := range testcases {
			sum += scoringstrategy.DecreaseScoreStrategy.TestcaseScore(tc)
		}

		item := scoringstrategy.DecreaseScoreStrategy.CalculateScore(testcases, 30)
		if sum != item {
			t.Errorf("expected per-testcase scores to sum to the item score %v, got %v", item, sum)
		}
	})
}
