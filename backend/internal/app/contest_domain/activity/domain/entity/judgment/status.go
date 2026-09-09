package judgment

var (
	QUEUED    = &Status{"Queued"}
	JUDGING   = &Status{"Judging"}
	JUDGED    = &Status{"Judged"}
	REJUDGED  = &Status{"Rejudged"}
	CANCELLED = &Status{"Cancelled"}
	ERROR     = &Status{"Error"}
)

type Status struct {
	value string
}

func (s Status) IsQueued() bool {
	return s.value == QUEUED.value
}

func (s Status) IsJudging() bool {
	return s.value == JUDGING.value
}

func (s Status) IsJudged() bool {
	return s.value == JUDGED.value
}

func (s Status) IsRejudged() bool {
	return s.value == REJUDGED.value
}

func (s Status) IsCancelled() bool {
	return s.value == CANCELLED.value
}

func (s Status) IsError() bool {
	return s.value == ERROR.value
}

func (s Status) String() string {
	return s.value
}
