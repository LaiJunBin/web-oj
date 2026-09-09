package submission_test

import (
	"judge-backend/internal/app/judge_domain/judge/domain/entity/submission"
	"testing"
	"time"
)

func TestSubmission(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		t.Run("should create submission with config values", func(t *testing.T) {
			s, err := submission.New(&submission.SubmissionConfig{
				UserID:           1,
				ContestProjectID: 10,
				GradingItemIds:   []int32{1, 2, 3},
				IsPartial:        true,
				Ignored:          false,
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if s.UserID() != 1 {
				t.Errorf("expected userID to be 1, got %d", s.UserID())
			}
			if s.ContestProjectID() != 10 {
				t.Errorf("expected contestProjectID to be 10, got %d", s.ContestProjectID())
			}
			if len(s.GradingItemIds()) != 3 {
				t.Errorf("expected 3 grading item IDs, got %d", len(s.GradingItemIds()))
			}
			if !s.IsPartial() {
				t.Errorf("expected isPartial to be true")
			}
			if s.IsIgnored() {
				t.Errorf("expected ignored to be false")
			}
		})

		t.Run("should default nil GradingItemIds to empty slice", func(t *testing.T) {
			s, err := submission.New(&submission.SubmissionConfig{
				UserID:           1,
				ContestProjectID: 10,
				GradingItemIds:   nil,
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if s.GradingItemIds() == nil {
				t.Errorf("expected GradingItemIds to be empty slice, got nil")
			}
			if len(s.GradingItemIds()) != 0 {
				t.Errorf("expected 0 grading item IDs, got %d", len(s.GradingItemIds()))
			}
		})

		t.Run("should use provided createdAt", func(t *testing.T) {
			fixedTime := time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC)
			s, err := submission.New(&submission.SubmissionConfig{
				UserID:           1,
				ContestProjectID: 10,
				CreatedAt:        fixedTime,
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !s.CreatedAt().Equal(fixedTime) {
				t.Errorf("expected createdAt to be %v, got %v", fixedTime, s.CreatedAt())
			}
		})

		t.Run("should default createdAt to now when zero", func(t *testing.T) {
			before := time.Now()
			s, err := submission.New(&submission.SubmissionConfig{
				UserID:           1,
				ContestProjectID: 10,
			})
			after := time.Now()

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if s.CreatedAt().Before(before) || s.CreatedAt().After(after) {
				t.Errorf("expected createdAt between %v and %v, got %v", before, after, s.CreatedAt())
			}
		})

		t.Run("should generate a unique ID at creation", func(t *testing.T) {
			s1, err := submission.New(&submission.SubmissionConfig{
				UserID:           1,
				ContestProjectID: 10,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			s2, err := submission.New(&submission.SubmissionConfig{
				UserID:           1,
				ContestProjectID: 10,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(s1.ID()) != 24 {
				t.Errorf("expected a 24-char hex ID, got %q", s1.ID())
			}
			if s1.ID() == s2.ID() {
				t.Errorf("expected unique IDs, both were %q", s1.ID())
			}
		})
	})

	t.Run("SetID", func(t *testing.T) {
		s, _ := submission.New(&submission.SubmissionConfig{
			UserID:           1,
			ContestProjectID: 10,
		})

		s.SetID("sub-123")
		if s.ID() != "sub-123" {
			t.Errorf("expected ID to be 'sub-123', got %s", s.ID())
		}
	})
}
