package judgestatus

var (
	QUEUED    = declareJudgeStatus("Queued")
	JUDGING   = declareJudgeStatus("Judging")
	JUDGED    = declareJudgeStatus("Judged")
	REJUDGED  = declareJudgeStatus("Rejudged")
	CANCELLED = declareJudgeStatus("Cancelled")
	ERROR     = declareJudgeStatus("Error")

	JudgeStatusMap = map[string]*JudgeStatus{}
)

type JudgeStatus struct {
	value string
}

func declareJudgeStatus(judgeStatus string) *JudgeStatus {
	j := &JudgeStatus{
		value: judgeStatus,
	}
	JudgeStatusMap[judgeStatus] = j
	return j
}

func (j JudgeStatus) String() string {
	return j.value
}

func (j JudgeStatus) IsQueued() bool {
	return j == *QUEUED
}

func (j JudgeStatus) IsJudging() bool {
	return j == *JUDGING
}

func (j JudgeStatus) IsJudged() bool {
	return j == *JUDGED
}

func (j JudgeStatus) IsRejudged() bool {
	return j == *REJUDGED
}

func (j JudgeStatus) IsCancelled() bool {
	return j == *CANCELLED
}

func (j JudgeStatus) IsError() bool {
	return j == *ERROR
}

// IsTerminal reports whether the status is one no further transition leaves:
// Judged, Rejudged, Cancelled or Error.
func (j JudgeStatus) IsTerminal() bool {
	return j.IsJudged() || j.IsRejudged() || j.IsCancelled() || j.IsError()
}
