package execresult_test

import (
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/execresult"
	"testing"
)

func TestExecResult(t *testing.T) {
	t.Run("NewExecResult", func(t *testing.T) {
		t.Run("should store all config values", func(t *testing.T) {
			result := execresult.NewExecResult(&execresult.ExecResultConfig{
				ExitCode: 0,
				ExecTime: 1500,
				Stdout:   "test passed",
				Stderr:   "",
			})

			if result.ExitCode() != 0 {
				t.Errorf("expected exitCode to be 0, got %d", result.ExitCode())
			}
			if result.ExecTime() != 1500 {
				t.Errorf("expected execTime to be 1500, got %d", result.ExecTime())
			}
			if result.Stdout() != "test passed" {
				t.Errorf("expected stdout to be 'test passed', got %s", result.Stdout())
			}
			if result.Stderr() != "" {
				t.Errorf("expected stderr to be empty, got %s", result.Stderr())
			}
		})

		t.Run("should handle non-zero exit code with stderr", func(t *testing.T) {
			result := execresult.NewExecResult(&execresult.ExecResultConfig{
				ExitCode: 1,
				ExecTime: 300,
				Stdout:   "",
				Stderr:   "segfault",
			})

			if result.ExitCode() != 1 {
				t.Errorf("expected exitCode to be 1, got %d", result.ExitCode())
			}
			if result.Stderr() != "segfault" {
				t.Errorf("expected stderr to be 'segfault', got %s", result.Stderr())
			}
		})
	})
}
