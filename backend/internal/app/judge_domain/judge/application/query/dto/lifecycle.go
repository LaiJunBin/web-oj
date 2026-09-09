package dto

type Lifecycle struct {
	Name                 string
	Command              string
	ExecTime             int64
	ExitCode             int32
	Stdout               string
	Stderr               string
	Status               string
	ExecutionEnvironment string
}
