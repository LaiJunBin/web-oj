package judgment_test

import (
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/gradingitem"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/gradingitem/scoringstrategy"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/judgestatus"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/judgetype"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/status"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/testcase"
	"testing"
	"time"
)

func newTestcase(id int32, score float64) *testcase.Testcase {
	return testcase.NewTestcase(&testcase.TestcaseConfig{
		ID:      id,
		Command: "echo hello",
		Score:   score,
	})
}

func newGradingItem(id int32, maxScore float64, tcs []*testcase.Testcase) *gradingitem.GradingItem {
	return gradingitem.NewGradingItem(&gradingitem.GradingItemConfig{
		ID:              id,
		Score:           maxScore,
		ScoringStrategy: scoringstrategy.IncreaseScoreStrategy,
		Testcases:       tcs,
	})
}

func TestJudgment(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		t.Run("should create judgment with default QUEUED status", func(t *testing.T) {
			j := judgment.New(&judgment.JudgmentConfig{
				SubmissionID: "sub-1",
				JudgeType:    judgetype.JUDGE,
			})

			if !j.Status().IsQueued() {
				t.Errorf("expected status to be Queued, got %s", j.Status().String())
			}
		})

		t.Run("should use custom status when provided", func(t *testing.T) {
			j := judgment.New(&judgment.JudgmentConfig{
				SubmissionID: "sub-1",
				JudgeType:    judgetype.JUDGE,
				Status:       judgestatus.JUDGING,
			})

			if !j.Status().IsJudging() {
				t.Errorf("expected status to be Judging, got %s", j.Status().String())
			}
		})

		t.Run("should calculate initial score from grading items max scores", func(t *testing.T) {
			tc1 := newTestcase(1, 30)
			tc2 := newTestcase(2, 70)
			gi1 := newGradingItem(1, 50, []*testcase.Testcase{tc1})
			gi2 := newGradingItem(2, 50, []*testcase.Testcase{tc2})

			j := judgment.New(&judgment.JudgmentConfig{
				SubmissionID: "sub-1",
				JudgeType:    judgetype.JUDGE,
				GradingItems: []*gradingitem.GradingItem{gi1, gi2},
			})

			// Initial score sums gradingItem.Score() (maxScore), not CalculateScore()
			if j.Score() != 100 {
				t.Errorf("expected initial score to be 100, got %f", j.Score())
			}
		})

		t.Run("should have zero score when no grading items", func(t *testing.T) {
			j := judgment.New(&judgment.JudgmentConfig{
				SubmissionID: "sub-1",
				JudgeType:    judgetype.JUDGE,
			})

			if j.Score() != 0 {
				t.Errorf("expected score to be 0, got %f", j.Score())
			}
		})

		t.Run("should use provided createdAt", func(t *testing.T) {
			fixedTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
			j := judgment.New(&judgment.JudgmentConfig{
				SubmissionID: "sub-1",
				JudgeType:    judgetype.JUDGE,
				CreatedAt:    fixedTime,
			})

			if !j.CreatedAt().Equal(fixedTime) {
				t.Errorf("expected createdAt to be %v, got %v", fixedTime, j.CreatedAt())
			}
		})

		t.Run("should default createdAt to now when zero", func(t *testing.T) {
			before := time.Now()
			j := judgment.New(&judgment.JudgmentConfig{
				SubmissionID: "sub-1",
				JudgeType:    judgetype.JUDGE,
			})
			after := time.Now()

			if j.CreatedAt().Before(before) || j.CreatedAt().After(after) {
				t.Errorf("expected createdAt to be between %v and %v, got %v", before, after, j.CreatedAt())
			}
		})

		t.Run("should store reason when provided", func(t *testing.T) {
			reason := "rejudge requested"
			j := judgment.New(&judgment.JudgmentConfig{
				SubmissionID: "sub-1",
				JudgeType:    judgetype.REJUDGE,
				Reason:       &reason,
			})

			if j.Reason() == nil || *j.Reason() != reason {
				t.Errorf("expected reason to be %q, got %v", reason, j.Reason())
			}
		})

		t.Run("should have nil reason when not provided", func(t *testing.T) {
			j := judgment.New(&judgment.JudgmentConfig{
				SubmissionID: "sub-1",
				JudgeType:    judgetype.JUDGE,
			})

			if j.Reason() != nil {
				t.Errorf("expected reason to be nil, got %v", j.Reason())
			}
		})
	})

	t.Run("CalculateScore", func(t *testing.T) {
		t.Run("should return zero with no grading items", func(t *testing.T) {
			j := judgment.New(&judgment.JudgmentConfig{
				SubmissionID: "sub-1",
				JudgeType:    judgetype.JUDGE,
			})

			if j.CalculateScore() != 0 {
				t.Errorf("expected calculated score to be 0, got %f", j.CalculateScore())
			}
		})

		t.Run("should sum calculated scores from all grading items", func(t *testing.T) {
			tc1 := newTestcase(1, 30)
			tc1.SetStatus(status.PASSED)
			tc1.SetParsedScore(30)
			tc2 := newTestcase(2, 70)
			tc2.SetStatus(status.PASSED)
			tc2.SetParsedScore(70)

			gi1 := newGradingItem(1, 50, []*testcase.Testcase{tc1})
			gi2 := newGradingItem(2, 50, []*testcase.Testcase{tc2})

			j := judgment.New(&judgment.JudgmentConfig{
				SubmissionID: "sub-1",
				JudgeType:    judgetype.JUDGE,
				GradingItems: []*gradingitem.GradingItem{gi1, gi2},
			})

			// increase strategy: sum of passed testcase scores, capped at maxScore
			// gi1: min(30, 50) = 30, gi2: min(70, 50) = 50
			score := j.CalculateScore()
			if score != 80 {
				t.Errorf("expected calculated score to be 80, got %f", score)
			}
		})

		t.Run("should return zero when no testcases passed", func(t *testing.T) {
			tc1 := newTestcase(1, 30)
			// testcase stays in PENDING status (default)
			gi := newGradingItem(1, 100, []*testcase.Testcase{tc1})

			j := judgment.New(&judgment.JudgmentConfig{
				SubmissionID: "sub-1",
				JudgeType:    judgetype.JUDGE,
				GradingItems: []*gradingitem.GradingItem{gi},
			})

			// increase strategy: no passed testcases → score = 0
			if j.CalculateScore() != 0 {
				t.Errorf("expected calculated score to be 0, got %f", j.CalculateScore())
			}
		})
	})

	t.Run("Setters", func(t *testing.T) {
		t.Run("SetID", func(t *testing.T) {
			j := judgment.New(&judgment.JudgmentConfig{
				SubmissionID: "sub-1",
				JudgeType:    judgetype.JUDGE,
			})

			j.SetID("judgment-123")
			if j.ID() != "judgment-123" {
				t.Errorf("expected ID to be 'judgment-123', got %s", j.ID())
			}
		})

		t.Run("SetStatus", func(t *testing.T) {
			j := judgment.New(&judgment.JudgmentConfig{
				SubmissionID: "sub-1",
				JudgeType:    judgetype.JUDGE,
			})

			j.SetStatus(judgestatus.JUDGING)
			if !j.Status().IsJudging() {
				t.Errorf("expected status to be Judging, got %s", j.Status().String())
			}

			j.SetStatus(judgestatus.JUDGED)
			if !j.Status().IsJudged() {
				t.Errorf("expected status to be Judged, got %s", j.Status().String())
			}

			j.SetStatus(judgestatus.ERROR)
			if !j.Status().IsError() {
				t.Errorf("expected status to be Error, got %s", j.Status().String())
			}

			j.SetStatus(judgestatus.CANCELLED)
			if !j.Status().IsCancelled() {
				t.Errorf("expected status to be Cancelled, got %s", j.Status().String())
			}

			j.SetStatus(judgestatus.REJUDGED)
			if !j.Status().IsRejudged() {
				t.Errorf("expected status to be Rejudged, got %s", j.Status().String())
			}
		})

		t.Run("SetRejudgeID", func(t *testing.T) {
			j := judgment.New(&judgment.JudgmentConfig{
				SubmissionID: "sub-1",
				JudgeType:    judgetype.JUDGE,
			})

			if j.RejudgeID() != nil {
				t.Errorf("expected rejudgeID to be nil initially")
			}

			j.SetRejudgeID("rejudge-1")
			if j.RejudgeID() == nil || *j.RejudgeID() != "rejudge-1" {
				t.Errorf("expected rejudgeID to be 'rejudge-1', got %v", j.RejudgeID())
			}
		})

		t.Run("SetScore", func(t *testing.T) {
			j := judgment.New(&judgment.JudgmentConfig{
				SubmissionID: "sub-1",
				JudgeType:    judgetype.JUDGE,
			})

			j.SetScore(95.5)
			if j.Score() != 95.5 {
				t.Errorf("expected score to be 95.5, got %f", j.Score())
			}
		})

		t.Run("SetErrorMessage", func(t *testing.T) {
			j := judgment.New(&judgment.JudgmentConfig{
				SubmissionID: "sub-1",
				JudgeType:    judgetype.JUDGE,
			})

			if j.ErrorMessage() != nil {
				t.Errorf("expected errorMessage to be nil initially")
			}

			j.SetErrorMessage("container timeout")
			if j.ErrorMessage() == nil || *j.ErrorMessage() != "container timeout" {
				t.Errorf("expected errorMessage to be 'container timeout', got %v", j.ErrorMessage())
			}
		})

		t.Run("SetReason", func(t *testing.T) {
			j := judgment.New(&judgment.JudgmentConfig{
				SubmissionID: "sub-1",
				JudgeType:    judgetype.JUDGE,
			})

			j.SetReason("manual rejudge")
			if j.Reason() == nil || *j.Reason() != "manual rejudge" {
				t.Errorf("expected reason to be 'manual rejudge', got %v", j.Reason())
			}
		})

		t.Run("SetSortingFields", func(t *testing.T) {
			j := judgment.New(&judgment.JudgmentConfig{
				SubmissionID: "sub-1",
				JudgeType:    judgetype.JUDGE,
			})

			fields := map[string]interface{}{"score": 100, "time": 500}
			j.SetSortingFields(fields)
			if j.SortingFields()["score"] != 100 {
				t.Errorf("expected sorting field 'score' to be 100, got %v", j.SortingFields()["score"])
			}
			if j.SortingFields()["time"] != 500 {
				t.Errorf("expected sorting field 'time' to be 500, got %v", j.SortingFields()["time"])
			}
		})
	})

	t.Run("Getters", func(t *testing.T) {
		t.Run("should return all config values", func(t *testing.T) {
			tc := newTestcase(1, 10)
			gi := newGradingItem(1, 100, []*testcase.Testcase{tc})

			j := judgment.New(&judgment.JudgmentConfig{
				SubmissionID: "sub-42",
				JudgeType:    judgetype.JUDGE_PARTIAL,
				GradingItems: []*gradingitem.GradingItem{gi},
			})

			if j.SubmissionID() != "sub-42" {
				t.Errorf("expected submissionID to be 'sub-42', got %s", j.SubmissionID())
			}
			if !j.JudgeType().IsJudgePartial() {
				t.Errorf("expected judgeType to be JUDGE_PARTIAL, got %s", j.JudgeType().String())
			}
			if len(j.GradingItems()) != 1 {
				t.Errorf("expected 1 grading item, got %d", len(j.GradingItems()))
			}
			if j.Project() != nil {
				t.Errorf("expected project to be nil when not provided")
			}
			if j.Lifecycle() != nil {
				t.Errorf("expected lifecycle to be nil when not provided")
			}
			if j.ErrorMessage() != nil {
				t.Errorf("expected errorMessage to be nil when not set")
			}
			if j.SortingFields() != nil {
				t.Errorf("expected sortingFields to be nil when not set")
			}
		})

		t.Run("should expose judging timestamps when provided", func(t *testing.T) {
			startedAt := time.Date(2024, 1, 1, 0, 0, 5, 0, time.UTC)
			finishedAt := time.Date(2024, 1, 1, 0, 0, 12, 0, time.UTC)

			j := judgment.New(&judgment.JudgmentConfig{
				SubmissionID:      "sub-42",
				JudgeType:         judgetype.JUDGE,
				JudgingStartedAt:  &startedAt,
				JudgingFinishedAt: &finishedAt,
			})

			if j.JudgingStartedAt() == nil || !j.JudgingStartedAt().Equal(startedAt) {
				t.Errorf("expected judgingStartedAt to be %v, got %v", startedAt, j.JudgingStartedAt())
			}
			if j.JudgingFinishedAt() == nil || !j.JudgingFinishedAt().Equal(finishedAt) {
				t.Errorf("expected judgingFinishedAt to be %v, got %v", finishedAt, j.JudgingFinishedAt())
			}
		})
	})

	t.Run("JudgingTimestampSetters", func(t *testing.T) {
		j := judgment.New(&judgment.JudgmentConfig{
			SubmissionID: "sub-1",
			JudgeType:    judgetype.JUDGE,
		})

		startedAt := time.Date(2024, 1, 1, 0, 0, 5, 0, time.UTC)
		finishedAt := time.Date(2024, 1, 1, 0, 0, 12, 0, time.UTC)

		j.SetJudgingStartedAt(startedAt)
		if j.JudgingStartedAt() == nil || !j.JudgingStartedAt().Equal(startedAt) {
			t.Errorf("expected judgingStartedAt to be %v, got %v", startedAt, j.JudgingStartedAt())
		}

		j.SetJudgingFinishedAt(finishedAt)
		if j.JudgingFinishedAt() == nil || !j.JudgingFinishedAt().Equal(finishedAt) {
			t.Errorf("expected judgingFinishedAt to be %v, got %v", finishedAt, j.JudgingFinishedAt())
		}
	})
}

func newQueuedJudgment(jt judgetype.JudgeType) *judgment.Judgment {
	return judgment.New(&judgment.JudgmentConfig{
		SubmissionID: "sub-1",
		JudgeType:    jt,
	})
}

func TestJudgmentTransitions(t *testing.T) {
	t.Run("Start", func(t *testing.T) {
		t.Run("should move QUEUED to JUDGING and stamp judgingStartedAt", func(t *testing.T) {
			j := newQueuedJudgment(judgetype.JUDGE)

			if err := j.Start(); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if !j.Status().IsJudging() {
				t.Errorf("expected status JUDGING, got %s", j.Status().String())
			}
			if j.JudgingStartedAt() == nil {
				t.Errorf("expected judgingStartedAt to be set")
			}
		})

		t.Run("should reject starting a non-queued judgment", func(t *testing.T) {
			j := newQueuedJudgment(judgetype.JUDGE)
			_ = j.Start()

			if err := j.Start(); err == nil {
				t.Errorf("expected error when starting a JUDGING judgment")
			}
		})
	})

	t.Run("Finish", func(t *testing.T) {
		t.Run("should move JUDGING to JUDGED for judge type", func(t *testing.T) {
			j := newQueuedJudgment(judgetype.JUDGE)
			_ = j.Start()

			if err := j.Finish(); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if !j.Status().IsJudged() {
				t.Errorf("expected status JUDGED, got %s", j.Status().String())
			}
			if j.JudgingFinishedAt() == nil {
				t.Errorf("expected judgingFinishedAt to be set")
			}
		})

		t.Run("should move JUDGING to REJUDGED for rejudge type", func(t *testing.T) {
			j := newQueuedJudgment(judgetype.REJUDGE)
			_ = j.Start()

			if err := j.Finish(); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if !j.Status().IsRejudged() {
				t.Errorf("expected status REJUDGED, got %s", j.Status().String())
			}
		})

		t.Run("should reject finishing a queued judgment", func(t *testing.T) {
			j := newQueuedJudgment(judgetype.JUDGE)

			if err := j.Finish(); err == nil {
				t.Errorf("expected error when finishing a QUEUED judgment")
			}
		})
	})

	t.Run("Cancel", func(t *testing.T) {
		t.Run("should cancel a queued judgment with reason", func(t *testing.T) {
			j := newQueuedJudgment(judgetype.JUDGE)

			if err := j.Cancel("wrong test data"); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if !j.Status().IsCancelled() {
				t.Errorf("expected status CANCELLED, got %s", j.Status().String())
			}
			if j.Reason() == nil || *j.Reason() != "wrong test data" {
				t.Errorf("expected reason to be set")
			}
			if j.JudgingFinishedAt() == nil {
				t.Errorf("expected judgingFinishedAt to be set")
			}
		})

		t.Run("should cancel a judging judgment and keep existing reason on empty input", func(t *testing.T) {
			j := newQueuedJudgment(judgetype.JUDGE)
			j.SetReason("original")
			_ = j.Start()

			if err := j.Cancel(""); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if j.Reason() == nil || *j.Reason() != "original" {
				t.Errorf("expected reason to stay %q, got %v", "original", j.Reason())
			}
		})

		t.Run("should reject cancelling a judged judgment", func(t *testing.T) {
			j := newQueuedJudgment(judgetype.JUDGE)
			_ = j.Start()
			_ = j.Finish()

			if err := j.Cancel("too late"); err == nil {
				t.Errorf("expected error when cancelling a JUDGED judgment")
			}
			if !j.Status().IsJudged() {
				t.Errorf("expected status to stay JUDGED, got %s", j.Status().String())
			}
		})

		t.Run("should reject cancelling an already cancelled judgment", func(t *testing.T) {
			j := newQueuedJudgment(judgetype.JUDGE)
			_ = j.Cancel("first")

			if err := j.Cancel("second"); err == nil {
				t.Errorf("expected error when cancelling twice")
			}
			if *j.Reason() != "first" {
				t.Errorf("expected reason to stay %q, got %q", "first", *j.Reason())
			}
		})
	})

	t.Run("Fail", func(t *testing.T) {
		t.Run("should fail a queued judgment", func(t *testing.T) {
			j := newQueuedJudgment(judgetype.JUDGE)

			if err := j.Fail(); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if !j.Status().IsError() {
				t.Errorf("expected status ERROR, got %s", j.Status().String())
			}
			if j.JudgingFinishedAt() == nil {
				t.Errorf("expected judgingFinishedAt to be set")
			}
		})

		t.Run("should fail a judging judgment", func(t *testing.T) {
			j := newQueuedJudgment(judgetype.JUDGE)
			_ = j.Start()

			if err := j.Fail(); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if !j.Status().IsError() {
				t.Errorf("expected status ERROR, got %s", j.Status().String())
			}
		})

		t.Run("should reject failing a terminal judgment", func(t *testing.T) {
			j := newQueuedJudgment(judgetype.JUDGE)
			_ = j.Start()
			_ = j.Finish()

			if err := j.Fail(); err == nil {
				t.Errorf("expected error when failing a JUDGED judgment")
			}
		})
	})
}
