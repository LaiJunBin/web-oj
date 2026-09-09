package lifecycle

import (
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/command"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/execresult"
)

var (
	project = ExecutionEnvironment{value: "project"}
	test    = ExecutionEnvironment{value: "test"}

	BeforeAll  = LifecycleType{value: "before_all"}
	BeforeEach = LifecycleType{value: "before_each"}
	AfterEach  = LifecycleType{value: "after_each"}
	AfterAll   = LifecycleType{value: "after_all"}
)

type ExecutionEnvironment struct {
	value string
}

func (e ExecutionEnvironment) IsProject() bool {
	return e.value == project.value
}

func (e ExecutionEnvironment) IsTest() bool {
	return e.value == test.value
}

func (e ExecutionEnvironment) String() string {
	return e.value
}

type LifecycleType struct {
	value string
}

func (t LifecycleType) String() string {
	return t.value
}

func (t LifecycleType) IsBeforeAll() bool {
	return t.value == BeforeAll.value
}

func (t LifecycleType) IsBeforeEach() bool {
	return t.value == BeforeEach.value
}

func (t LifecycleType) IsAfterEach() bool {
	return t.value == AfterEach.value
}

func (t LifecycleType) IsAfterAll() bool {
	return t.value == AfterAll.value
}

type Lifecycle struct {
	id                   int32
	executionEnvironment ExecutionEnvironment
	_type                LifecycleType
	command              command.Command
	result               *execresult.ExecResult
}

type LifecycleConfig struct {
	ID                   int32
	ExecutionEnvironment string
	Type                 string
	Command              string
}

func NewLifecycle(config *LifecycleConfig) *Lifecycle {
	return &Lifecycle{
		id:                   config.ID,
		executionEnvironment: ExecutionEnvironment{value: config.ExecutionEnvironment},
		_type:                LifecycleType{value: config.Type},
		command:              command.New(config.Command),
	}
}

func (l *Lifecycle) ID() int32 {
	return l.id
}

func (l *Lifecycle) ExecutionEnvironment() ExecutionEnvironment {
	return l.executionEnvironment
}

func (l *Lifecycle) Type() LifecycleType {
	return l._type
}

func (l *Lifecycle) Command() command.Command {
	return l.command
}

func (l *Lifecycle) Result() *execresult.ExecResult {
	return l.result
}

func (l *Lifecycle) SetResult(result *execresult.ExecResult) {
	l.result = result
}
