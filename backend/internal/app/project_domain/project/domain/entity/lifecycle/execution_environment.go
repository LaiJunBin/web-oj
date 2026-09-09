package lifecycle

import "errors"

var (
	ExecutionEnvironmentProject = declareExecutionEnvironment("project")
	ExecutionEnvironmentTest    = declareExecutionEnvironment("test")

	ErrInvalidExecutionEnvironment = errors.New("invalid execution environment")

	ExecutionEnvironments = map[string]ExecutionEnvironment{
		ExecutionEnvironmentProject.value: ExecutionEnvironmentProject,
		ExecutionEnvironmentTest.value:    ExecutionEnvironmentTest,
	}
)

type ExecutionEnvironment struct {
	value string
}

var executionEnvironments = []ExecutionEnvironment{}

func GetExecutionEnvironments() []ExecutionEnvironment {
	return executionEnvironments
}

func newExecutionEnvironment(value string) (ExecutionEnvironment, error) {
	env := ExecutionEnvironment{value: value}
	if err := env.validate(); err != nil {
		return ExecutionEnvironment{}, err
	}
	return env, nil
}

func mustNewExecutionEnvironment(value string) ExecutionEnvironment {
	env, err := newExecutionEnvironment(value)
	if err != nil {
		panic(err)
	}
	return env
}

func declareExecutionEnvironment(value string) ExecutionEnvironment {
	env := mustNewExecutionEnvironment(value)
	executionEnvironments = append(executionEnvironments, env)
	return env
}

func (e ExecutionEnvironment) validate() error {
	if e.value == "" {
		return ErrInvalidExecutionEnvironment
	}
	return nil
}

func (e ExecutionEnvironment) String() string {
	return e.value
}

func (e ExecutionEnvironment) Equals(other ExecutionEnvironment) bool {
	return e.value == other.value
}

func (e ExecutionEnvironment) IsProject() bool {
	return e.Equals(ExecutionEnvironmentProject)
}

func (e ExecutionEnvironment) IsTest() bool {
	return e.Equals(ExecutionEnvironmentTest)
}
