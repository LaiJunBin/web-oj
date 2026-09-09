package dto

import (
	"time"
)

type JudgeWorkerInfo struct {
	ID              string
	Status          string
	JudgeOrder      []string
	CreatedAt       time.Time
	LastExecutionAt *time.Time
}

type Log struct {
	Text string
	Time time.Time
}

type JudgeWorker struct {
	ID              string
	Status          string
	JudgeOrder      []string
	CreatedAt       time.Time
	LastExecutionAt *time.Time
	Logs            []*Log
}
