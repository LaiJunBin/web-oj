package dto

import "time"

type Lifecycle struct {
	BeforeAll  []*Command
	BeforeEach []*Command
	AfterEach  []*Command
	AfterAll   []*Command
}

type LifecycleCommand struct {
	ID                   int32
	Type                 string
	ExecutionEnvironment string
	Name                 string
	Description          string
	Command              string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}
