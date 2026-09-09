package status

var (
	PENDING   = declareStatus("Pending")
	RUNNING   = declareStatus("Running")
	PASSED    = declareStatus("Passed")
	FAILED    = declareStatus("Failed")
	ERROR     = declareStatus("Error")
	TIMEOUT   = declareStatus("Timeout")
	CANCELLED = declareStatus("Cancelled")

	StatusMap = map[string]Status{}
)

type Status struct {
	value string
}

func declareStatus(status string) Status {
	StatusMap[status] = Status{status}
	return StatusMap[status]
}

func (s Status) IsPending() bool {
	return s.value == PENDING.value
}

func (s Status) IsRunning() bool {
	return s.value == RUNNING.value
}

func (s Status) IsPassed() bool {
	return s.value == PASSED.value
}

func (s Status) IsFailed() bool {
	return s.value == FAILED.value
}

func (s Status) IsError() bool {
	return s.value == ERROR.value
}

func (s Status) IsTimeout() bool {
	return s.value == TIMEOUT.value
}

func (s Status) IsCancelled() bool {
	return s.value == CANCELLED.value
}

func (s Status) String() string {
	return s.value
}
