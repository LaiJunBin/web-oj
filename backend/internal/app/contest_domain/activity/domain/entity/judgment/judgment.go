package judgment

type Judgment struct {
	id           string
	submissionID string
	status       Status
}

func New(id string, submissionID string, status string) *Judgment {
	return &Judgment{
		id:           id,
		submissionID: submissionID,
		status:       Status{value: status},
	}
}

func (j *Judgment) ID() string {
	return j.id
}

func (j *Judgment) SubmissionID() string {
	return j.submissionID
}

func (j *Judgment) Status() Status {
	return j.status
}

func (j *Judgment) CanCancel() bool {
	return j.status.IsQueued() || j.status.IsJudging()
}
