package getsystemdefaults_test

import (
	"reflect"
	"sort"
	"testing"

	getsystemdefaults "judge-backend/internal/app/judge_domain/worker/application/usecase/get_system_defaults"
	"judge-backend/internal/config"
)

func TestExecuteReportsTheValuesInForce(t *testing.T) {
	usecase := getsystemdefaults.NewUsecase(&config.Config{
		Judge: &config.JudgeConfig{
			MaxJudgeDurationMinutes:   45,
			SubmissionCooldownSeconds: 30,
			PidsLimit:                 intPtr(512),
			MemoryLimitMB:             intPtr(2048),
		},
	})

	out, err := usecase.Execute(&getsystemdefaults.Input{})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	if out.MaxJudgeDurationMinutes != 45 {
		t.Errorf("MaxJudgeDurationMinutes = %d, want 45", out.MaxJudgeDurationMinutes)
	}
	if out.SubmissionCooldownSeconds != 30 {
		t.Errorf("SubmissionCooldownSeconds = %d, want 30", out.SubmissionCooldownSeconds)
	}
	if out.PidsLimit != 512 {
		t.Errorf("PidsLimit = %d, want 512", out.PidsLimit)
	}
	if out.MemoryLimitMB != 2048 {
		t.Errorf("MemoryLimitMB = %d, want 2048", out.MemoryLimitMB)
	}
}

// The shipped memory default. The admin UI renders this as "預設(不限制)", so
// reporting anything else would describe a cap that is not there.
func TestExecuteReportsAnAbsentMemoryLimit(t *testing.T) {
	usecase := getsystemdefaults.NewUsecase(&config.Config{
		Judge: &config.JudgeConfig{},
	})

	out, _ := usecase.Execute(&getsystemdefaults.Input{})

	if out.MemoryLimitMB != 0 {
		t.Errorf("MemoryLimitMB = %d, want 0", out.MemoryLimitMB)
	}
}

func intPtr(v int) *int { return &v }

// An explicit 0 in config.yaml is an operator switching the cap off, and the
// admin UI has to render that as "預設(不限制)" rather than as an unset field.
func TestExecuteReportsAnExplicitZeroLimit(t *testing.T) {
	usecase := getsystemdefaults.NewUsecase(&config.Config{
		Judge: &config.JudgeConfig{PidsLimit: intPtr(0)},
	})

	out, _ := usecase.Execute(&getsystemdefaults.Input{})

	if out.PidsLimit != 0 {
		t.Errorf("PidsLimit = %d, want 0 (unlimited)", out.PidsLimit)
	}
}

// A negative value is a third spelling of "off" that config.yaml still
// accepts. Reported verbatim it would have the admin UI offer "預設(-1)" as a
// number to copy.
func TestExecuteReportsANegativeLimitAsUnlimited(t *testing.T) {
	usecase := getsystemdefaults.NewUsecase(&config.Config{
		Judge: &config.JudgeConfig{PidsLimit: intPtr(-1), MemoryLimitMB: intPtr(-1)},
	})

	out, _ := usecase.Execute(&getsystemdefaults.Input{})

	if out.PidsLimit != 0 {
		t.Errorf("PidsLimit = %d, want 0 (unlimited)", out.PidsLimit)
	}
	if out.MemoryLimitMB != 0 {
		t.Errorf("MemoryLimitMB = %d, want 0 (unlimited)", out.MemoryLimitMB)
	}
}

// Zero is a real value for the cooldown — "this deployment runs without one" —
// so it must be reported rather than treated as unset.
func TestExecuteReportsAZeroCooldown(t *testing.T) {
	usecase := getsystemdefaults.NewUsecase(&config.Config{
		Judge: &config.JudgeConfig{MaxJudgeDurationMinutes: 30, SubmissionCooldownSeconds: 0},
	})

	out, _ := usecase.Execute(&getsystemdefaults.Input{})

	if out.SubmissionCooldownSeconds != 0 {
		t.Errorf("SubmissionCooldownSeconds = %d, want 0", out.SubmissionCooldownSeconds)
	}
}

// This endpoint reads from the same struct that holds the app/JWT/CSRF keys,
// the database and Redis passwords, the Sonar token and the internal docker
// host. The allowlist is the only thing keeping those out of an
// authenticated-user-readable response, so widening it has to be deliberate:
// this test fails the moment a field is added or renamed.
func TestSystemDefaultsExposesOnlyKnownFields(t *testing.T) {
	want := []string{
		"MaxJudgeDurationMinutes",
		"SubmissionCooldownSeconds",
		"PidsLimit",
		"MemoryLimitMB",
	}

	typ := reflect.TypeOf(getsystemdefaults.Output{})
	got := make([]string, 0, typ.NumField())
	for i := 0; i < typ.NumField(); i++ {
		got = append(got, typ.Field(i).Name)
	}

	sort.Strings(got)
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("system defaults expose %v, want %v — if this is intentional, confirm the new field "+
			"is not a secret and update this list", got, want)
	}
}

// Unlike maxJudgeDurationMinutes, the cooldown has no fallback in
// config.applyDefaults, so a negative reaches this endpoint. It is admitted as
// "no cooldown" at the enforcement point, so reporting it verbatim describes a
// state the deployment is not in — and put 預設(-1) in the admin form, which
// is what the frontend team hit.
func TestExecuteReportsANegativeCooldownAsUnlimited(t *testing.T) {
	usecase := getsystemdefaults.NewUsecase(&config.Config{
		Judge: &config.JudgeConfig{SubmissionCooldownSeconds: -5},
	})

	out, _ := usecase.Execute(&getsystemdefaults.Input{})

	if out.SubmissionCooldownSeconds != 0 {
		t.Errorf("SubmissionCooldownSeconds = %d, want 0 (no cooldown)", out.SubmissionCooldownSeconds)
	}
}
