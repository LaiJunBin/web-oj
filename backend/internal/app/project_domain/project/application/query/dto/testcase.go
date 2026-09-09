package dto

import "time"

type TestcaseInfo struct {
	ID        int32
	Name      string
	Score     float64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Testcase struct {
	TestcaseInfo TestcaseInfo
	IsSample     bool
	Command      string
}

type DependencyTestcaseInfo struct {
	ID   int32
	Name string
}

type DependencyTestcase struct {
	Testcase    DependencyTestcaseInfo
	GradingItem GradingItemInfo
}

type TestcaseDetail struct {
	Testcase
	Dependencies []*DependencyTestcase
}

type ProjectTestcaseInfo struct {
	ID   int32
	Name string
}

type ProjectTestcaseGradingItem struct {
	ID        int32
	Label     string
	Name      string
	Testcases []*ProjectTestcaseInfo
}
