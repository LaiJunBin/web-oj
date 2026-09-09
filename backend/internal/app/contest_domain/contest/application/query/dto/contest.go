package dto

import "time"

type Contest struct {
	ID   int32
	Name string
}

type ContestDetail struct {
	Contest
	Projects []Project
}

type ContestScheduleProject struct {
	Name        string
	Description string
	StartTime   time.Time
	EndTime     time.Time
	Score       float64
}

type ContestSchedule struct {
	IsPublic bool
	Projects []*ContestScheduleProject
}
