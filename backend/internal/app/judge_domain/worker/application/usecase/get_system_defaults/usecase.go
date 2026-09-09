package getsystemdefaults

import (
	"judge-backend/internal/config"
)

type Input struct {
}

// Output is the deployment's answer to "what happens when a contest project
// leaves this field blank?".
//
// It is a HAND-WRITTEN allowlist and must stay one. The configuration it
// reads from also holds the app/JWT/CSRF/session keys, the database and Redis
// passwords, the Sonar token and the internal docker host — so this must
// never become a projection of config.Config, and must never be filled by
// reflection over it. Adding a field here is a deliberate act; TestSystemDefaultsExposesOnlyKnownFields
// fails if the shape changes without one.
type Output struct {
	MaxJudgeDurationMinutes   int32
	SubmissionCooldownSeconds int32
	PidsLimit                 int32
	MemoryLimitMB             int32
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	config *config.Config
}

func NewUsecase(
	config *config.Config,
) Usecase {
	return &usecase{
		config: config,
	}
}

// Execute reports the values actually in force, after the config package has
// applied its fallbacks — an admin needs the number a blank field resolves
// to, not what happens to be written in the file.
func (u *usecase) Execute(input *Input) (*Output, error) {
	return &Output{
		MaxJudgeDurationMinutes: int32(u.config.Judge.MaxJudgeDurationMinutes),
		// No fallback of its own: zero is a real value here (no cooldown),
		// so config leaves it as supplied.
		SubmissionCooldownSeconds: reportedCap(u.config.Judge.SubmissionCooldownSeconds),
		PidsLimit:                 reportedOptionalCap(u.config.Judge.PidsLimit),
		MemoryLimitMB:             reportedOptionalCap(u.config.Judge.MemoryLimitMB),
	}, nil
}

// reportedCap spells a disabled limit the way the API does — 0 — rather than
// echoing the configuration verbatim.
//
// Every non-positive value behaves identically where the limit is enforced:
// SubmissionCooldown.Reserve admits immediately, and the runtime sets no cap.
// So a negative in config.yaml is a real state with a real meaning, and
// reporting it as-is describes the deployment accurately only in the sense
// that a wrong number is a number. This endpoint exists to tell an
// administrator what to type into a blank field, and 預設(-1) is not something
// anybody can act on.
func reportedCap(configured int) int32 {
	if configured < 0 {
		return 0
	}
	return int32(configured)
}

// reportedOptionalCap is reportedCap for the keys that may be absent, where
// absence means the built-in default is off. Only the memory limit reaches
// here as nil; config.applyDefaults has already filled the pids limit.
func reportedOptionalCap(configured *int) int32 {
	if configured == nil {
		return 0
	}
	return reportedCap(*configured)
}
