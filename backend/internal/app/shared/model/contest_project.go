package model

import "time"

type ContestProject struct {
	ID                    int32 `gorm:"primaryKey;autoIncrement"`
	ContestID             int32 `gorm:"index"`
	ProjectID             int32 `gorm:"index"`
	Name                  string
	ActivateTime          time.Time
	DeactivateTime        time.Time
	StartTime             time.Time
	EndTime               time.Time
	IsPublic              bool
	JudgePriorityStrategy string
	// MaxJudgeDurationMinutes caps one judgment's total wall-clock for this
	// contest project; nil falls back to the global judge config.
	MaxJudgeDurationMinutes *int32
	// SubmissionCooldownSeconds is the minimum gap between one contestant's
	// submissions to this contest project. nil falls back to the global
	// config; 0 disables the wait entirely.
	SubmissionCooldownSeconds *int32
	// PidsLimit caps the processes a judging container for this contest
	// project may create. nil falls back to the global config; 0 removes the
	// cap. It is the one limit MaxJudgeDurationMinutes cannot substitute for:
	// a fork bomb exhausts the host's process table in seconds, long before
	// the judgment timer fires and while the daemon is already too starved to
	// answer the request that would tear the container down.
	PidsLimit *int32
	// MemoryLimitMB caps a judging container's memory. nil falls back to the
	// global config; 0 removes the cap. Its value is containment rather than
	// thrift: without it the kernel picks the OOM victim host-wide, so one
	// contestant's runaway allocation can take down another judgment, MySQL,
	// or the backend. With it the kill lands inside the offending cgroup and
	// the failure is attributable to the submission that caused it.
	MemoryLimitMB *int32
	// ScoreboardCompact switches the public/contestant scoreboard to the
	// compact rendering (single bar per row); admin views are unaffected.
	ScoreboardCompact bool
	// ScoreboardMaskNames masks contestant display names on the
	// public/contestant scoreboard; admin views are unaffected.
	ScoreboardMaskNames bool

	// relations
	Contest Contest `gorm:"foreignKey:ContestID"`
	Project Project `gorm:"foreignKey:ProjectID"`
}

func (cp *ContestProject) IsActivated() bool {
	return cp.ActivateTime.Before(time.Now()) && cp.DeactivateTime.After(time.Now()) && cp.Contest.IsActivated()
}

func (cp *ContestProject) TimeLeft() time.Duration {
	if cp.IsEnded() {
		return 0
	}

	if cp.IsStarted() {
		return time.Until(cp.EndTime)
	}

	return time.Since(cp.StartTime)
}

func (p *ContestProject) IsStarted() bool {
	return p.StartTime.Before(time.Now())
}

func (p *ContestProject) IsEnded() bool {
	return p.EndTime.Before(time.Now())
}

func (p *ContestProject) Status() string {
	if p.IsEnded() {
		return "Ended"
	}

	if p.IsStarted() {
		return "InProgress"
	}

	return "NotStarted"
}
