package submission

var (
	QUEUED    = declareStatus("Queued")
	JUDGING   = declareStatus("Judging")
	JUDGED    = declareStatus("Judged")
	REJUDGED  = declareStatus("Rejudged")
	CANCELLED = declareStatus("Cancelled")
	ERROR     = declareStatus("Error")

	StatusMap = map[string]*Status{}
)

type Status struct {
	value string
}

func declareStatus(status string) *Status {
	j := &Status{
		value: status,
	}
	StatusMap[status] = j
	return j
}

func (j Status) String() string {
	return j.value
}

func (j Status) IsQueued() bool {
	return j == *QUEUED
}

func (j Status) IsJudging() bool {
	return j == *JUDGING
}

func (j Status) IsJudged() bool {
	return j == *JUDGED
}

func (j Status) IsRejudged() bool {
	return j == *REJUDGED
}

func (j Status) IsCancelled() bool {
	return j == *CANCELLED
}

func (j Status) IsError() bool {
	return j == *ERROR
}
