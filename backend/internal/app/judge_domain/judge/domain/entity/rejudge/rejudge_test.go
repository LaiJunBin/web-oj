package rejudge_test

import (
	"judge-backend/internal/app/judge_domain/judge/domain/entity/rejudge"
	"testing"
	"time"
)

func TestRejudge(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		t.Run("should create rejudge with config values", func(t *testing.T) {
			r := rejudge.New(&rejudge.RejudgeConfig{
				ID:     "rejudge-1",
				Reason: "wrong answer on testcase 3",
				Status: rejudge.PENDING,
			})

			if r.ID() != "rejudge-1" {
				t.Errorf("expected ID to be 'rejudge-1', got %s", r.ID())
			}
			if r.Reason() != "wrong answer on testcase 3" {
				t.Errorf("expected reason to be 'wrong answer on testcase 3', got %s", r.Reason())
			}
			if !r.Status().IsPending() {
				t.Errorf("expected status to be Pending, got %s", r.Status().String())
			}
		})

		// The config carried a contest project id long before the entity had a
		// field to put it in, so New silently dropped it and every rejudge was
		// stored without the project it belonged to.
		t.Run("should keep the contest project id from the config", func(t *testing.T) {
			r := rejudge.New(&rejudge.RejudgeConfig{
				ID:               "rejudge-1",
				ContestProjectID: 42,
				Reason:           "test",
				Status:           rejudge.PENDING,
			})

			if r.ContestProjectID() != 42 {
				t.Errorf("expected contestProjectID to be 42, got %d", r.ContestProjectID())
			}
		})

		t.Run("should use provided createdAt", func(t *testing.T) {
			fixedTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
			r := rejudge.New(&rejudge.RejudgeConfig{
				ID:        "rejudge-1",
				Reason:    "test",
				Status:    rejudge.PENDING,
				CreatedAt: fixedTime,
			})

			if !r.CreatedAt().Equal(fixedTime) {
				t.Errorf("expected createdAt to be %v, got %v", fixedTime, r.CreatedAt())
			}
		})

		t.Run("should default createdAt to now when zero", func(t *testing.T) {
			before := time.Now()
			r := rejudge.New(&rejudge.RejudgeConfig{
				ID:     "rejudge-1",
				Reason: "test",
				Status: rejudge.PENDING,
			})
			after := time.Now()

			if r.CreatedAt().Before(before) || r.CreatedAt().After(after) {
				t.Errorf("expected createdAt between %v and %v, got %v", before, after, r.CreatedAt())
			}
		})

		t.Run("should initialize with provided judgments", func(t *testing.T) {
			r := rejudge.New(&rejudge.RejudgeConfig{
				ID:        "rejudge-1",
				Reason:    "test",
				Status:    rejudge.PENDING,
				Judgments: []string{"j-1", "j-2"},
			})

			if len(r.Judgments()) != 2 {
				t.Errorf("expected 2 judgments, got %d", len(r.Judgments()))
			}
		})
	})

	t.Run("SetStatus", func(t *testing.T) {
		r := rejudge.New(&rejudge.RejudgeConfig{
			ID:     "rejudge-1",
			Reason: "test",
			Status: rejudge.PENDING,
		})

		r.SetStatus(rejudge.RUNNING)
		if !r.Status().IsRunning() {
			t.Errorf("expected status to be Running, got %s", r.Status().String())
		}

		r.SetStatus(rejudge.FINISHED)
		if !r.Status().IsFinished() {
			t.Errorf("expected status to be Finished, got %s", r.Status().String())
		}

		r.SetStatus(rejudge.APPROVED)
		if !r.Status().IsApproved() {
			t.Errorf("expected status to be Approved, got %s", r.Status().String())
		}

		r.SetStatus(rejudge.REJECTED)
		if !r.Status().IsRejected() {
			t.Errorf("expected status to be Rejected, got %s", r.Status().String())
		}

		r.SetStatus(rejudge.ERROR)
		if !r.Status().IsError() {
			t.Errorf("expected status to be Error, got %s", r.Status().String())
		}

		r.SetStatus(rejudge.CANCELED)
		if !r.Status().IsCanceled() {
			t.Errorf("expected status to be Canceled, got %s", r.Status().String())
		}
	})

	t.Run("AddJudgmentID", func(t *testing.T) {
		r := rejudge.New(&rejudge.RejudgeConfig{
			ID:     "rejudge-1",
			Reason: "test",
			Status: rejudge.PENDING,
		})

		if len(r.Judgments()) != 0 {
			t.Errorf("expected 0 judgments initially, got %d", len(r.Judgments()))
		}

		r.AddJudgmentID("judgment-1")
		if len(r.Judgments()) != 1 {
			t.Errorf("expected 1 judgment, got %d", len(r.Judgments()))
		}
		if r.Judgments()[0] != "judgment-1" {
			t.Errorf("expected judgment[0] to be 'judgment-1', got %s", r.Judgments()[0])
		}

		r.AddJudgmentID("judgment-2")
		if len(r.Judgments()) != 2 {
			t.Errorf("expected 2 judgments, got %d", len(r.Judgments()))
		}
	})
}

func TestRejudgeStatus(t *testing.T) {
	t.Run("NewStatus", func(t *testing.T) {
		validStatuses := []struct {
			value    string
			expected *rejudge.Status
		}{
			{"Pending", rejudge.PENDING},
			{"Running", rejudge.RUNNING},
			{"Finished", rejudge.FINISHED},
			{"Approved", rejudge.APPROVED},
			{"Rejected", rejudge.REJECTED},
			{"Error", rejudge.ERROR},
			{"Canceled", rejudge.CANCELED},
		}

		for _, tt := range validStatuses {
			t.Run("should create "+tt.value+" status", func(t *testing.T) {
				s, err := rejudge.NewStatus(tt.value)
				if err != nil {
					t.Errorf("unexpected error for %s: %v", tt.value, err)
				}
				if s != tt.expected {
					t.Errorf("expected %v, got %v", tt.expected, s)
				}
			})
		}

		t.Run("should return error for invalid status", func(t *testing.T) {
			_, err := rejudge.NewStatus("invalid")
			if err == nil {
				t.Errorf("expected error for invalid status, got nil")
			}
		})

		t.Run("should return error for empty status", func(t *testing.T) {
			_, err := rejudge.NewStatus("")
			if err == nil {
				t.Errorf("expected error for empty status, got nil")
			}
		})
	})
}
