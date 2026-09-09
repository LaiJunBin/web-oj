package dto

import "time"

type AdminContest struct {
	ID                    int32
	Name                  string
	Description           string
	IsPublic              bool
	TotalProjects         int32
	ActivationTime        time.Time
	DeactivationTime      time.Time
	JudgePriorityStrategy string
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type AdminContestDetail struct {
	ID               int32
	Name             string
	Description      string
	IsPublic         bool
	Projects         []*AdminContestProject
	ActivationTime   time.Time
	DeactivationTime time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type AdminContestProject struct {
	ID                        int32
	ProjectID                 int32
	Name                      string
	Description               string
	IsPublic                  bool
	Image                     *string
	TestImage                 *string
	ActivateTime              time.Time
	DeactivateTime            time.Time
	StartTime                 time.Time
	EndTime                   time.Time
	JudgePriorityStrategy     string
	MaxJudgeDurationMinutes   *int32
	SubmissionCooldownSeconds *int32
	PidsLimit                 *int32
	MemoryLimitMB             *int32
	ScoreboardCompact         bool
	ScoreboardMaskNames       bool
}
