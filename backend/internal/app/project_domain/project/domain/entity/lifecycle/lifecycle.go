package lifecycle

import (
	"errors"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/project_domain/project/domain/entity/lifecycle/command"
	"judge-backend/internal/app/project_domain/project/domain/entity/project"
)

var (
	BEFORE_ALL  = "before_all"
	BEFORE_EACH = "before_each"
	AFTER_EACH  = "after_each"
	AFTER_ALL   = "after_all"

	Types = map[string]struct{}{
		BEFORE_ALL:  {},
		BEFORE_EACH: {},
		AFTER_EACH:  {},
		AFTER_ALL:   {},
	}

	ErrorInvalidType                 = errors.New("invalid type")
	ErrorInvalidExecutionEnvironment = errors.New("invalid execution environment")
)

type Lifecycle struct {
	id                   ID
	projectID            project.ID
	executionEnvironment ExecutionEnvironment
	lifecycleType        string
	name                 Name
	description          string
	command              command.Command

	sharedtypes.TimeStamps
}

type LifecycleConfig struct {
	ProjectID            project.ID
	ExecutionEnvironment string
	Type                 string
	Name                 string
	Description          string
	Command              string
}

func New(config *LifecycleConfig) (*Lifecycle, error) {
	if _, ok := Types[config.Type]; !ok {
		return nil, ErrorInvalidType
	}

	execEnv, ok := ExecutionEnvironments[config.ExecutionEnvironment]
	if !ok {
		return nil, ErrorInvalidExecutionEnvironment
	}

	name, err := NewName(config.Name)
	if err != nil {
		return nil, err
	}

	command, err := command.New(config.Command)
	if err != nil {
		return nil, err
	}

	return &Lifecycle{
		projectID:            config.ProjectID,
		executionEnvironment: execEnv,
		lifecycleType:        config.Type,
		name:                 name,
		description:          config.Description,
		command:              command,
	}, nil
}

func (l *Lifecycle) ID() ID {
	return l.id
}

func (l *Lifecycle) ProjectID() project.ID {
	return l.projectID
}

func (l *Lifecycle) ExecutionEnvironment() ExecutionEnvironment {
	return l.executionEnvironment
}

func (l *Lifecycle) Type() string {
	return l.lifecycleType
}

func (l *Lifecycle) Name() Name {
	return l.name
}

func (l *Lifecycle) Description() string {
	return l.description
}

func (l *Lifecycle) Command() command.Command {
	return l.command
}

func (l *Lifecycle) SetID(id ID) {
	l.id = id
}

func (l *Lifecycle) SetExecutionEnvironment(ee string) error {
	e, ok := ExecutionEnvironments[ee]
	if !ok {
		return ErrorInvalidExecutionEnvironment
	}

	l.executionEnvironment = e
	return nil
}

func (l *Lifecycle) SetType(lifecycleType string) error {
	if _, ok := Types[lifecycleType]; !ok {
		return ErrorInvalidType
	}

	l.lifecycleType = lifecycleType
	return nil
}

func (l *Lifecycle) SetName(name string) error {
	n, err := NewName(name)
	if err != nil {
		return err
	}

	l.name = n
	return nil
}

func (l *Lifecycle) SetDescription(description string) {
	l.description = description
}

func (l *Lifecycle) SetCommand(cmd string) error {
	c, err := command.New(cmd)
	if err != nil {
		return err
	}

	l.command = c
	return nil
}
