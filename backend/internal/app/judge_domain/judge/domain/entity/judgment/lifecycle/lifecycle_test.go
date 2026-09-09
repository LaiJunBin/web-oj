package lifecycle_test

import (
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/execresult"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/lifecycle"
	"testing"
)

func TestLifecycle(t *testing.T) {
	t.Run("NewLifecycle", func(t *testing.T) {
		t.Run("should create lifecycle with config values", func(t *testing.T) {
			lc := lifecycle.NewLifecycle(&lifecycle.LifecycleConfig{
				ID:                   1,
				ExecutionEnvironment: "project",
				Type:                 "before_all",
				Command:              "make setup",
			})

			if lc.ID() != 1 {
				t.Errorf("expected ID to be 1, got %d", lc.ID())
			}
			if !lc.ExecutionEnvironment().IsProject() {
				t.Errorf("expected execution environment to be project")
			}
			if !lc.Type().IsBeforeAll() {
				t.Errorf("expected type to be before_all, got %s", lc.Type().String())
			}
			if lc.Command().String() != "make setup" {
				t.Errorf("expected command to be 'make setup', got %s", lc.Command().String())
			}
			if lc.Result() != nil {
				t.Errorf("expected result to be nil initially")
			}
		})

		t.Run("should handle test execution environment", func(t *testing.T) {
			lc := lifecycle.NewLifecycle(&lifecycle.LifecycleConfig{
				ID:                   2,
				ExecutionEnvironment: "test",
				Type:                 "after_each",
				Command:              "cleanup.sh",
			})

			if !lc.ExecutionEnvironment().IsTest() {
				t.Errorf("expected execution environment to be test")
			}
			if lc.ExecutionEnvironment().IsProject() {
				t.Errorf("expected execution environment to NOT be project")
			}
		})
	})

	t.Run("SetResult", func(t *testing.T) {
		lc := lifecycle.NewLifecycle(&lifecycle.LifecycleConfig{
			ID:                   1,
			ExecutionEnvironment: "project",
			Type:                 "before_all",
			Command:              "make setup",
		})

		result := execresult.NewExecResult(&execresult.ExecResultConfig{
			ExitCode: 0,
			ExecTime: 100,
			Stdout:   "done",
			Stderr:   "",
		})

		lc.SetResult(result)
		if lc.Result() == nil {
			t.Errorf("expected result to be set")
		}
		if lc.Result().ExitCode() != 0 {
			t.Errorf("expected result exitCode to be 0, got %d", lc.Result().ExitCode())
		}
	})

	t.Run("LifecycleType", func(t *testing.T) {
		types := []struct {
			name   string
			check  func(lifecycle.LifecycleType) bool
			others []func(lifecycle.LifecycleType) bool
		}{
			{
				name:  "before_all",
				check: lifecycle.LifecycleType.IsBeforeAll,
				others: []func(lifecycle.LifecycleType) bool{
					lifecycle.LifecycleType.IsBeforeEach,
					lifecycle.LifecycleType.IsAfterEach,
					lifecycle.LifecycleType.IsAfterAll,
				},
			},
			{
				name:  "before_each",
				check: lifecycle.LifecycleType.IsBeforeEach,
				others: []func(lifecycle.LifecycleType) bool{
					lifecycle.LifecycleType.IsBeforeAll,
					lifecycle.LifecycleType.IsAfterEach,
					lifecycle.LifecycleType.IsAfterAll,
				},
			},
			{
				name:  "after_each",
				check: lifecycle.LifecycleType.IsAfterEach,
				others: []func(lifecycle.LifecycleType) bool{
					lifecycle.LifecycleType.IsBeforeAll,
					lifecycle.LifecycleType.IsBeforeEach,
					lifecycle.LifecycleType.IsAfterAll,
				},
			},
			{
				name:  "after_all",
				check: lifecycle.LifecycleType.IsAfterAll,
				others: []func(lifecycle.LifecycleType) bool{
					lifecycle.LifecycleType.IsBeforeAll,
					lifecycle.LifecycleType.IsBeforeEach,
					lifecycle.LifecycleType.IsAfterEach,
				},
			},
		}

		for _, tt := range types {
			t.Run(tt.name, func(t *testing.T) {
				lc := lifecycle.NewLifecycle(&lifecycle.LifecycleConfig{
					ID:                   1,
					ExecutionEnvironment: "project",
					Type:                 tt.name,
					Command:              "test",
				})

				if !tt.check(lc.Type()) {
					t.Errorf("expected Is%s to be true", tt.name)
				}
				for _, other := range tt.others {
					if other(lc.Type()) {
						t.Errorf("expected other type checks to be false for %s", tt.name)
					}
				}
			})
		}
	})
}

func TestJudgmentLifecycle(t *testing.T) {
	t.Run("should filter only BeforeAll and AfterAll lifecycles", func(t *testing.T) {
		lifecycles := []*lifecycle.Lifecycle{
			lifecycle.NewLifecycle(&lifecycle.LifecycleConfig{
				ID: 1, ExecutionEnvironment: "project", Type: "before_all", Command: "setup",
			}),
			lifecycle.NewLifecycle(&lifecycle.LifecycleConfig{
				ID: 2, ExecutionEnvironment: "test", Type: "before_each", Command: "pre-test",
			}),
			lifecycle.NewLifecycle(&lifecycle.LifecycleConfig{
				ID: 3, ExecutionEnvironment: "project", Type: "after_each", Command: "post-test",
			}),
			lifecycle.NewLifecycle(&lifecycle.LifecycleConfig{
				ID: 4, ExecutionEnvironment: "project", Type: "after_all", Command: "teardown",
			}),
		}

		jl := lifecycle.NewJudgmentLifecycle(lifecycles)

		if len(jl.BeforeAll()) != 1 {
			t.Errorf("expected 1 BeforeAll lifecycle, got %d", len(jl.BeforeAll()))
		}
		if len(jl.AfterAll()) != 1 {
			t.Errorf("expected 1 AfterAll lifecycle, got %d", len(jl.AfterAll()))
		}
		if jl.BeforeAll()[0].Command().String() != "setup" {
			t.Errorf("expected BeforeAll command to be 'setup', got %s", jl.BeforeAll()[0].Command().String())
		}
		if jl.AfterAll()[0].Command().String() != "teardown" {
			t.Errorf("expected AfterAll command to be 'teardown', got %s", jl.AfterAll()[0].Command().String())
		}
	})

	t.Run("should return empty slices when no matching types", func(t *testing.T) {
		lifecycles := []*lifecycle.Lifecycle{
			lifecycle.NewLifecycle(&lifecycle.LifecycleConfig{
				ID: 1, ExecutionEnvironment: "test", Type: "before_each", Command: "pre",
			}),
			lifecycle.NewLifecycle(&lifecycle.LifecycleConfig{
				ID: 2, ExecutionEnvironment: "test", Type: "after_each", Command: "post",
			}),
		}

		jl := lifecycle.NewJudgmentLifecycle(lifecycles)

		if len(jl.BeforeAll()) != 0 {
			t.Errorf("expected 0 BeforeAll lifecycles, got %d", len(jl.BeforeAll()))
		}
		if len(jl.AfterAll()) != 0 {
			t.Errorf("expected 0 AfterAll lifecycles, got %d", len(jl.AfterAll()))
		}
	})

	t.Run("should handle empty input", func(t *testing.T) {
		jl := lifecycle.NewJudgmentLifecycle(nil)

		if len(jl.BeforeAll()) != 0 {
			t.Errorf("expected 0 BeforeAll lifecycles, got %d", len(jl.BeforeAll()))
		}
		if len(jl.AfterAll()) != 0 {
			t.Errorf("expected 0 AfterAll lifecycles, got %d", len(jl.AfterAll()))
		}
	})

	t.Run("should handle multiple BeforeAll lifecycles", func(t *testing.T) {
		lifecycles := []*lifecycle.Lifecycle{
			lifecycle.NewLifecycle(&lifecycle.LifecycleConfig{
				ID: 1, ExecutionEnvironment: "project", Type: "before_all", Command: "setup1",
			}),
			lifecycle.NewLifecycle(&lifecycle.LifecycleConfig{
				ID: 2, ExecutionEnvironment: "project", Type: "before_all", Command: "setup2",
			}),
		}

		jl := lifecycle.NewJudgmentLifecycle(lifecycles)

		if len(jl.BeforeAll()) != 2 {
			t.Errorf("expected 2 BeforeAll lifecycles, got %d", len(jl.BeforeAll()))
		}
	})
}

func TestTestcaseLifecycle(t *testing.T) {
	t.Run("should filter only BeforeEach and AfterEach lifecycles", func(t *testing.T) {
		lifecycles := []*lifecycle.Lifecycle{
			lifecycle.NewLifecycle(&lifecycle.LifecycleConfig{
				ID: 1, ExecutionEnvironment: "project", Type: "before_all", Command: "setup",
			}),
			lifecycle.NewLifecycle(&lifecycle.LifecycleConfig{
				ID: 2, ExecutionEnvironment: "test", Type: "before_each", Command: "pre-test",
			}),
			lifecycle.NewLifecycle(&lifecycle.LifecycleConfig{
				ID: 3, ExecutionEnvironment: "project", Type: "after_each", Command: "post-test",
			}),
			lifecycle.NewLifecycle(&lifecycle.LifecycleConfig{
				ID: 4, ExecutionEnvironment: "project", Type: "after_all", Command: "teardown",
			}),
		}

		tl := lifecycle.NewTestcaseLifecycle(lifecycles)

		if len(tl.BeforeEach()) != 1 {
			t.Errorf("expected 1 BeforeEach lifecycle, got %d", len(tl.BeforeEach()))
		}
		if len(tl.AfterEach()) != 1 {
			t.Errorf("expected 1 AfterEach lifecycle, got %d", len(tl.AfterEach()))
		}
		if tl.BeforeEach()[0].Command().String() != "pre-test" {
			t.Errorf("expected BeforeEach command to be 'pre-test', got %s", tl.BeforeEach()[0].Command().String())
		}
		if tl.AfterEach()[0].Command().String() != "post-test" {
			t.Errorf("expected AfterEach command to be 'post-test', got %s", tl.AfterEach()[0].Command().String())
		}
	})

	t.Run("should return empty slices when no matching types", func(t *testing.T) {
		lifecycles := []*lifecycle.Lifecycle{
			lifecycle.NewLifecycle(&lifecycle.LifecycleConfig{
				ID: 1, ExecutionEnvironment: "project", Type: "before_all", Command: "setup",
			}),
		}

		tl := lifecycle.NewTestcaseLifecycle(lifecycles)

		if len(tl.BeforeEach()) != 0 {
			t.Errorf("expected 0 BeforeEach lifecycles, got %d", len(tl.BeforeEach()))
		}
		if len(tl.AfterEach()) != 0 {
			t.Errorf("expected 0 AfterEach lifecycles, got %d", len(tl.AfterEach()))
		}
	})

	t.Run("should handle empty input", func(t *testing.T) {
		tl := lifecycle.NewTestcaseLifecycle(nil)

		if len(tl.BeforeEach()) != 0 {
			t.Errorf("expected 0 BeforeEach lifecycles, got %d", len(tl.BeforeEach()))
		}
		if len(tl.AfterEach()) != 0 {
			t.Errorf("expected 0 AfterEach lifecycles, got %d", len(tl.AfterEach()))
		}
	})
}
