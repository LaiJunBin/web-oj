package dto

import "time"

type ContestProject struct {
	ID                        int32
	ProjectID                 int32
	Name                      string
	Description               string
	IsPublic                  bool
	JudgePriorityStrategy     string
	Image                     *string
	TestImage                 *string
	ActivateTime              time.Time
	DeactivateTime            time.Time
	StartTime                 time.Time
	EndTime                   time.Time
	MaxJudgeDurationMinutes   *int32
	SubmissionCooldownSeconds *int32
	PidsLimit                 *int32
	MemoryLimitMB             *int32
	ScoreboardCompact         bool
	ScoreboardMaskNames       bool
}

type UserScore struct {
	User         *User
	Score        float64
	SubmissionID string
}

type ScoreDetailItem struct {
	GradingItemID int32
	Label         string
	Name          string
	Description   string
	Score         float64
	TotalScore    float64
	Status        string
}

type ContestantScoreDetail struct {
	User  *User
	Items []*ScoreDetailItem
}
