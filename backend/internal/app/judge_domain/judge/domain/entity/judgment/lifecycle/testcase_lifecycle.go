package lifecycle

type TestcaseLifecycle struct {
	beforeEach []Lifecycle
	afterEach  []Lifecycle
}

func NewTestcaseLifecycle(lifecycles []*Lifecycle) *TestcaseLifecycle {
	beforeEach := make([]Lifecycle, 0)
	afterEach := make([]Lifecycle, 0)
	for _, l := range lifecycles {
		switch true {
		case l.Type().IsBeforeEach():
			beforeEach = append(beforeEach, Lifecycle{
				id:                   l.ID(),
				_type:                l.Type(),
				executionEnvironment: l.ExecutionEnvironment(),
				command:              l.Command(),
			})
		case l.Type().IsAfterEach():
			afterEach = append(afterEach, Lifecycle{
				id:                   l.ID(),
				_type:                l.Type(),
				executionEnvironment: l.ExecutionEnvironment(),
				command:              l.Command(),
			})
		}
	}

	return &TestcaseLifecycle{
		beforeEach: beforeEach,
		afterEach:  afterEach,
	}
}

func (l *TestcaseLifecycle) BeforeEach() []Lifecycle {
	return l.beforeEach
}

func (l *TestcaseLifecycle) AfterEach() []Lifecycle {
	return l.afterEach
}
