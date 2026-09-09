package testcase

import (
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/command"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/execresult"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/lifecycle"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/status"
)

type Testcase struct {
	id              int32
	name            string
	command         command.Command
	configuredScore float64
	parsedScore     float64
	dependTestcases []int32
	status          status.Status
	lifecycle       *lifecycle.TestcaseLifecycle
	result          *execresult.ExecResult
}

type TestcaseConfig struct {
	ID              int32
	Name            string
	Command         string
	Score           float64
	ParsedScore     float64
	DependTestcases []int32
}

func NewTestcase(config *TestcaseConfig) *Testcase {
	return &Testcase{
		id:              config.ID,
		name:            config.Name,
		command:         command.New(config.Command),
		configuredScore: config.Score,
		parsedScore:     config.ParsedScore,
		dependTestcases: config.DependTestcases,
		status:          status.PENDING,
	}
}

func (t *Testcase) ID() int32 {
	return t.id
}

// Name is carried purely so a judgment can record which testcase it ran under
// the name it had at the time; nothing in the judging logic reads it.
func (t *Testcase) Name() string {
	return t.name
}

func (t *Testcase) Command() command.Command {
	return t.command
}

// Score returns the configured max score for this testcase.
func (t *Testcase) Score() float64 {
	return t.configuredScore
}

// ParsedScore returns the score set by the result parser after execution.
func (t *Testcase) ParsedScore() float64 {
	return t.parsedScore
}

func (t *Testcase) Result() *execresult.ExecResult {
	return t.result
}

func (t *Testcase) Status() status.Status {
	return t.status
}

func (t *Testcase) Lifecycle() *lifecycle.TestcaseLifecycle {
	return t.lifecycle
}

func (t *Testcase) SetParsedScore(score float64) {
	t.parsedScore = score
}

func (t *Testcase) SetResult(result *execresult.ExecResult) {
	t.result = result
}

func (t *Testcase) SetStatus(status status.Status) {
	t.status = status
}

func (t *Testcase) SetLifecycle(lifecycle *lifecycle.TestcaseLifecycle) {
	t.lifecycle = lifecycle
}

func (t *Testcase) DependencyMet(passedTestcases map[int32]struct{}) bool {
	for _, id := range t.dependTestcases {
		if _, ok := passedTestcases[id]; !ok {
			return false
		}
	}
	return true
}
