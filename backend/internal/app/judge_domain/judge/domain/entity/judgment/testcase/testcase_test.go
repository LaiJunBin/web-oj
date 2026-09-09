package testcase_test

import (
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/status"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/testcase"
	"testing"
)

func TestTestcase(t *testing.T) {
	t.Run("NewTestcase", func(t *testing.T) {
		t.Run("should create testcase with PENDING status", func(t *testing.T) {
			tc := testcase.NewTestcase(&testcase.TestcaseConfig{
				ID:      1,
				Command: "python3 test.py",
				Score:   10,
			})

			if !tc.Status().IsPending() {
				t.Errorf("expected status to be Pending, got %s", tc.Status().String())
			}
		})

		t.Run("should store config values", func(t *testing.T) {
			tc := testcase.NewTestcase(&testcase.TestcaseConfig{
				ID:              5,
				Command:         "go test ./...",
				Score:           25.5,
				DependTestcases: []int32{1, 2, 3},
			})

			if tc.ID() != 5 {
				t.Errorf("expected ID to be 5, got %d", tc.ID())
			}
			if tc.Command().String() != "go test ./..." {
				t.Errorf("expected command to be 'go test ./...', got %s", tc.Command().String())
			}
			if tc.Score() != 25.5 {
				t.Errorf("expected score to be 25.5, got %f", tc.Score())
			}
		})

		t.Run("should have nil result and lifecycle initially", func(t *testing.T) {
			tc := testcase.NewTestcase(&testcase.TestcaseConfig{
				ID:      1,
				Command: "echo test",
				Score:   10,
			})

			if tc.Result() != nil {
				t.Errorf("expected result to be nil initially")
			}
			if tc.Lifecycle() != nil {
				t.Errorf("expected lifecycle to be nil initially")
			}
		})
	})

	t.Run("DependencyMet", func(t *testing.T) {
		t.Run("should return true when no dependencies", func(t *testing.T) {
			tc := testcase.NewTestcase(&testcase.TestcaseConfig{
				ID:      1,
				Command: "echo test",
				Score:   10,
			})

			if !tc.DependencyMet(map[int32]struct{}{}) {
				t.Errorf("expected DependencyMet to be true with no dependencies")
			}
		})

		t.Run("should return true when all dependencies are met", func(t *testing.T) {
			tc := testcase.NewTestcase(&testcase.TestcaseConfig{
				ID:              3,
				Command:         "echo test",
				Score:           10,
				DependTestcases: []int32{1, 2},
			})

			passed := map[int32]struct{}{1: {}, 2: {}}
			if !tc.DependencyMet(passed) {
				t.Errorf("expected DependencyMet to be true when all deps are met")
			}
		})

		t.Run("should return false when some dependencies are not met", func(t *testing.T) {
			tc := testcase.NewTestcase(&testcase.TestcaseConfig{
				ID:              3,
				Command:         "echo test",
				Score:           10,
				DependTestcases: []int32{1, 2},
			})

			passed := map[int32]struct{}{1: {}}
			if tc.DependencyMet(passed) {
				t.Errorf("expected DependencyMet to be false when dep 2 is missing")
			}
		})

		t.Run("should return false when no dependencies are met", func(t *testing.T) {
			tc := testcase.NewTestcase(&testcase.TestcaseConfig{
				ID:              3,
				Command:         "echo test",
				Score:           10,
				DependTestcases: []int32{1, 2},
			})

			if tc.DependencyMet(map[int32]struct{}{}) {
				t.Errorf("expected DependencyMet to be false when no deps are met")
			}
		})
	})

	t.Run("Setters", func(t *testing.T) {
		t.Run("SetStatus", func(t *testing.T) {
			tc := testcase.NewTestcase(&testcase.TestcaseConfig{
				ID:      1,
				Command: "echo test",
				Score:   10,
			})

			tc.SetStatus(status.RUNNING)
			if !tc.Status().IsRunning() {
				t.Errorf("expected status to be Running, got %s", tc.Status().String())
			}

			tc.SetStatus(status.PASSED)
			if !tc.Status().IsPassed() {
				t.Errorf("expected status to be Passed, got %s", tc.Status().String())
			}

			tc.SetStatus(status.FAILED)
			if !tc.Status().IsFailed() {
				t.Errorf("expected status to be Failed, got %s", tc.Status().String())
			}

			tc.SetStatus(status.ERROR)
			if !tc.Status().IsError() {
				t.Errorf("expected status to be Error, got %s", tc.Status().String())
			}

			tc.SetStatus(status.TIMEOUT)
			if !tc.Status().IsTimeout() {
				t.Errorf("expected status to be Timeout, got %s", tc.Status().String())
			}

			tc.SetStatus(status.CANCELLED)
			if !tc.Status().IsCancelled() {
				t.Errorf("expected status to be Cancelled, got %s", tc.Status().String())
			}
		})

		t.Run("SetParsedScore", func(t *testing.T) {
			tc := testcase.NewTestcase(&testcase.TestcaseConfig{
				ID:      1,
				Command: "echo test",
				Score:   10,
			})

			tc.SetParsedScore(8)
			if tc.ParsedScore() != 8 {
				t.Errorf("expected parsed score to be 8, got %f", tc.ParsedScore())
			}
			if tc.Score() != 10 {
				t.Errorf("expected configured score to remain 10, got %f", tc.Score())
			}
		})
	})
}
