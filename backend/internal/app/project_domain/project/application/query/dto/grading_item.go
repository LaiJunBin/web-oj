package dto

import "time"

type GradingItemInfo struct {
	ID    int32
	Label string
	Name  string
}

type GradingItem struct {
	ID          int32
	Label       string
	Name        string
	Description string
	Score       float64
	Order       uint32
	TestcaseLen int32
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type GradingItemDetail struct {
	GradingItem
	ScoringStrategy string
	OutputFormat    string
	ResultParser    string
	Dependencies    []*DependencyTestcase
	Testcases       []*Testcase
}
