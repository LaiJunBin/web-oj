package prioritystrategy

import "fmt"

// JudgePriorityStrategy is an enum of available judge queue priority strategies.
// To add a new strategy: add a const below, register it in judge_domain/judge/domain/prioritystrategies/,
// and add it to the all slice.
type JudgePriorityStrategy string

const (
	Default       JudgePriorityStrategy = "default"
	SubmitTime    JudgePriorityStrategy = "submit_time"
	TimeSinceLast JudgePriorityStrategy = "time_since_last"
)

var all = []JudgePriorityStrategy{Default, SubmitTime, TimeSinceLast}

// All returns all registered priority strategies.
func All() []JudgePriorityStrategy {
	return all
}

// New creates a JudgePriorityStrategy from a string, returning an error if the value is not registered.
func New(value string) (JudgePriorityStrategy, error) {
	s := JudgePriorityStrategy(value)
	if !s.Valid() {
		return "", fmt.Errorf("invalid judge priority strategy: %s", value)
	}
	return s, nil
}

func (s JudgePriorityStrategy) String() string {
	return string(s)
}

func (s JudgePriorityStrategy) Valid() bool {
	for _, v := range all {
		if v == s {
			return true
		}
	}
	return false
}
