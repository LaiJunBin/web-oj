package lifecycle

type JudgmentLifecycle struct {
	beforeAll []Lifecycle
	afterAll  []Lifecycle
}

func NewJudgmentLifecycle(lifecycles []*Lifecycle) *JudgmentLifecycle {
	beforeAll := make([]Lifecycle, 0)
	afterAll := make([]Lifecycle, 0)
	for _, l := range lifecycles {
		switch true {
		case l.Type().IsBeforeAll():
			beforeAll = append(beforeAll, Lifecycle{
				id:                   l.ID(),
				_type:                l.Type(),
				executionEnvironment: l.ExecutionEnvironment(),
				command:              l.Command(),
			})
		case l.Type().IsAfterAll():
			afterAll = append(afterAll, Lifecycle{
				id:                   l.ID(),
				_type:                l.Type(),
				executionEnvironment: l.ExecutionEnvironment(),
				command:              l.Command(),
			})
		}
	}

	return &JudgmentLifecycle{
		beforeAll: beforeAll,
		afterAll:  afterAll,
	}
}

func (l *JudgmentLifecycle) BeforeAll() []Lifecycle {
	return l.beforeAll
}

func (l *JudgmentLifecycle) AfterAll() []Lifecycle {
	return l.afterAll
}
