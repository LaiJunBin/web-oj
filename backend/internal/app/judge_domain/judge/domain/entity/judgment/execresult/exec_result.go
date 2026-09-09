package execresult

type ExecResult struct {
	exitCode int
	execTime int64
	stdout   string
	stderr   string
}

type ExecResultConfig struct {
	ExitCode int
	ExecTime int64
	Stdout   string
	Stderr   string
}

func NewExecResult(config *ExecResultConfig) *ExecResult {
	return &ExecResult{
		exitCode: config.ExitCode,
		execTime: config.ExecTime,
		stdout:   config.Stdout,
		stderr:   config.Stderr,
	}
}

func (e *ExecResult) ExitCode() int {
	return e.exitCode
}

func (e *ExecResult) ExecTime() int64 {
	return e.execTime
}

func (e *ExecResult) Stdout() string {
	return e.stdout
}

func (e *ExecResult) Stderr() string {
	return e.stderr
}
