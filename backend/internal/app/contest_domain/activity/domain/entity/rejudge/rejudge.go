package rejudge

type Rejudge struct {
	id        string
	reason    string
	judgments []string
	status    *Status
}

type RejudgeConfig struct {
	ID        string
	Reason    string
	Judgments []string
	Status    string
}

func New(
	config *RejudgeConfig,
) (*Rejudge, error) {
	status, err := getStatus(config.Status)
	if err != nil {
		return nil, err
	}

	return &Rejudge{
		id:        config.ID,
		reason:    config.Reason,
		judgments: config.Judgments,
		status:    status,
	}, nil
}

func (r *Rejudge) ID() string {
	return r.id
}

func (r *Rejudge) Status() *Status {
	return r.status
}

func (r *Rejudge) Reason() string {
	return r.reason
}

func (r *Rejudge) Judgments() []string {
	return r.judgments
}

func (r *Rejudge) CanCancel() bool {
	return r.status.IsPending() || r.status.IsRunning()
}

func (r *Rejudge) CanApprove() bool {
	return r.status.IsFinished()
}

func (r *Rejudge) CanReject() bool {
	return r.status.IsFinished()
}
