package gradingitem

import (
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/execresult"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/gradingitem/scoringstrategy"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/status"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/testcase"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/testcase/resultparser"
)

type GradingItem struct {
	id              int32
	label           string
	name            string
	score           float64
	scoringStrategy scoringstrategy.ScoringStrategy
	dependTestcases []int32
	testcases       []*testcase.Testcase
	resultParser    resultparser.ResultParser
}

type GradingItemConfig struct {
	ID              int32
	Label           string
	Name            string
	Score           float64
	ScoringStrategy scoringstrategy.ScoringStrategy
	DependTestcases []int32
	Testcases       []*testcase.Testcase
	ResultParser    resultparser.ResultParser
}

func NewGradingItem(config *GradingItemConfig) *GradingItem {
	parser := config.ResultParser
	if parser == nil {
		parser = resultparser.AutoParser
	}
	return &GradingItem{
		id:              config.ID,
		label:           config.Label,
		name:            config.Name,
		score:           config.Score,
		scoringStrategy: config.ScoringStrategy,
		dependTestcases: config.DependTestcases,
		testcases:       config.Testcases,
		resultParser:    parser,
	}
}

func (t *GradingItem) ID() int32 {
	return t.id
}

// Label and Name are carried purely so a judgment can record which grading item
// it scored under the name it had at the time; nothing in the judging logic
// reads them.
func (t *GradingItem) Label() string {
	return t.label
}

func (t *GradingItem) Name() string {
	return t.name
}

func (t *GradingItem) Score() float64 {
	return t.score
}

func (t *GradingItem) CalculateScore() float64 {
	return t.scoringStrategy.CalculateScore(t.testcases, t.score)
}

func (t *GradingItem) ScoringStrategy() scoringstrategy.ScoringStrategy {
	return t.scoringStrategy
}

func (t *GradingItem) Status() status.Status {
	if len(t.testcases) == 0 {
		return status.PASSED
	}

	pendingCount := 0
	hasRunning := false
	hasFailed := false

	for _, tc := range t.testcases {
		tcStatus := tc.Status()

		if tcStatus.IsPending() {
			pendingCount++
			continue
		}

		if tcStatus.IsRunning() {
			hasRunning = true
			continue
		}

		if !tcStatus.IsPassed() {
			hasFailed = true
		}
	}

	if pendingCount == len(t.testcases) {
		return status.PENDING
	}

	if pendingCount > 0 || hasRunning {
		return status.RUNNING
	}

	if hasFailed {
		return status.FAILED
	}

	return status.PASSED
}

func (t *GradingItem) SetScore(score float64) {
	t.score = score
}

func (t *GradingItem) Testcases() []*testcase.Testcase {
	return t.testcases
}

func (t *GradingItem) DependencyMet(passedTestcases map[int32]struct{}) bool {
	for _, id := range t.dependTestcases {
		if _, ok := passedTestcases[id]; !ok {
			return false
		}
	}
	return true
}

func (t *GradingItem) ResultParser() resultparser.ResultParser {
	return t.resultParser
}

func (t *GradingItem) ParseResult(result *execresult.ExecResult, configuredScore float64) (status.Status, float64, string) {
	return t.resultParser.Parse(result, configuredScore)
}
