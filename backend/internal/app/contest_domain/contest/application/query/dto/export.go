package dto

import "time"

// ContestExport is a contest loaded for writing to a bundle.
//
// Separate from AdminContestDetail because it needs one thing that DTO has no
// reason to carry: the NAME of the underlying problem. The screen shows the
// slot's own name and never needs the library entry's, but a bundle references
// problems by name, so without it the export has nothing to write.
type ContestExport struct {
	ID             int32
	Name           string
	Description    string
	IsPublic       bool
	ActivateTime   time.Time
	DeactivateTime time.Time
	Projects       []ContestProjectExport
}

type ContestProjectExport struct {
	// ProjectName is the library entry this slot points at.
	ProjectName string

	Name                      string
	IsPublic                  bool
	JudgePriorityStrategy     string
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
