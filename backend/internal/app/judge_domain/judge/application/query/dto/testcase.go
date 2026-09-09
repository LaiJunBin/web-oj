package dto

type TestCase struct {
	ID                  int32
	Name                string
	Command             string
	ExecTime            int64
	ExitCode            int32
	IsSample            bool
	Stdout              string
	Stderr              string
	Status              string
	BeforeEachLifecycle []Lifecycle
	AfterEachLifecycle  []Lifecycle
}
