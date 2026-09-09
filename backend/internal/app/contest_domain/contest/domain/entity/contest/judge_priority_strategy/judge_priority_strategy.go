package judgeprioritystrategy

type JudgePriorityStrategy struct {
	value string
}

func newJudgePriorityStrategy(value string) JudgePriorityStrategy {
	return JudgePriorityStrategy{value: value}
}

// NewFromValidated creates a JudgePriorityStrategy from a pre-validated strategy name.
// Callers are responsible for ensuring the name is valid before calling this constructor.
func NewFromValidated(value string) JudgePriorityStrategy {
	return newJudgePriorityStrategy(value)
}

func NewDefault() JudgePriorityStrategy {
	return newJudgePriorityStrategy("default")
}

func (j JudgePriorityStrategy) String() string {
	return j.value
}

func (j JudgePriorityStrategy) Valid() bool {
	return j.value != ""
}
