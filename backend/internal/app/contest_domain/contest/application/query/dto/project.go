package dto

type Project struct {
	ID     int32
	Name   string
	Status string
}

type ProjectDetail struct {
	Project
	Description   string
	DynamicFields []DynamicField
	GradingItems  []GradingItem
	Assets        []AssetInfo
	TimeLeft      int64
}

// UserProjectDetail is the user-facing variant of ProjectDetail; the user and
// public contracts diverge on participant-awareness, so they must not share
// one DTO.
type UserProjectDetail struct {
	ProjectDetail
	IsParticipant bool
}

type DynamicField struct {
	Label        string
	Name         string
	Type         string
	IsRequired   bool
	DefaultValue *string
}

type GradingItem struct {
	ID              int32
	Name            string
	Label           string
	Description     string
	Score           float64
	OutputFormat    string
	PublicTestcases []Testcase
}

type Testcase struct {
	ID   int32
	Name string
}
