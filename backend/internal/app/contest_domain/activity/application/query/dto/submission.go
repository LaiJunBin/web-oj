package dto

import "time"

type ContestInfo struct {
	ID   int32
	Name string
}

type ProjectInfo struct {
	ID   int32
	Name string
}

type Submission struct {
	ID         string
	Contest    *ContestInfo
	Project    *ProjectInfo
	User       *User
	IsPartial  bool
	Ignored    bool
	Type       string
	Status     string
	Score      float64
	TotalScore float64
	Time       time.Time
}

type Project struct {
	ID           int32
	ImageTag     string
	TestImageTag string
	Env          []string
	TestEnv      []string
	MountPath    string
	ResultParser string
}

type Result struct {
	ExitCode int32
	ExecTime int64
	Stdout   string
	Stderr   string
}

type Lifecycle struct {
	ID                   int32
	ExecutionEnvironment string
	Command              string
	Result               Result
}

type JudgmentLifeycle struct {
	BeforeAll []*Lifecycle
	AfterAll  []*Lifecycle
}

type TestcaseLifecycle struct {
	BeforeEach []*Lifecycle
	AfterEach  []*Lifecycle
}

// Testcase and GradingItem decode straight out of a judgment's stored snapshot,
// so their keys are spelled out rather than left to the driver's default
// lowercasing — the document uses camelCase, and a silent mismatch here shows up
// as a zero score rather than an error.
//
// Score is what the testcase earned and TotalScore its configured max, matching
// the grading item and judgment levels. ParsedScore is the raw parser output the
// earned score derives from.
type Testcase struct {
	ID          int32              `bson:"id"`
	Name        string             `bson:"name"`
	Score       float64            `bson:"score"`
	TotalScore  float64            `bson:"totalScore"`
	ParsedScore float64            `bson:"parsedScore"`
	Status      string             `bson:"status"`
	Command     string             `bson:"command"`
	Lifecycle   *TestcaseLifecycle `bson:"lifecycle"`
	Result      Result             `bson:"result"`
}

type GradingItem struct {
	ID              int32       `bson:"id"`
	Label           string      `bson:"label"`
	Name            string      `bson:"name"`
	Score           float64     `bson:"score"`
	TotalScore      float64     `bson:"totalScore"`
	Status          string      `bson:"status"`
	ScoringStrategy string      `bson:"scoringStrategy"`
	ResultParser    string      `bson:"resultParser"`
	Testcases       []*Testcase `bson:"testcases"`
}

type Judgment struct {
	ID                string
	JudgeType         string
	Project           *Project
	GradingItems      []*GradingItem
	Lifecycle         *JudgmentLifeycle
	Score             float64
	TotalScore        float64
	Status            string
	RejudgeID         *string
	CreatedAt         time.Time
	JudgingStartedAt  *time.Time
	JudgingFinishedAt *time.Time
	// ErrorMessage carries the reason a judgment ended in Error (e.g.
	// provisioning retries exhausted). Admin-facing only — contestant APIs
	// must never expose it, as it may reference infrastructure internals.
	ErrorMessage *string
	// StaffNote is the operator-written, contestant-visible note; nil when
	// none was written (or it was cleared). The admin detail echoes it so
	// the edit form shows the current value.
	StaffNote *string
}

type TestcaseInfo struct {
	ID    int32
	Name  string
	Score float64
}

type GradingItemInfo struct {
	ID        int32
	Name      string
	Score     float64
	Testcases []*TestcaseInfo
}

type SubmissionDetail struct {
	ID           string
	User         *User
	Contest      *ContestInfo
	Project      *ProjectInfo
	GradingItems []*GradingItemInfo
	IsPartial    bool
	Ignored      bool
	Judgments    []*Judgment
	CreatedAt    time.Time
}
