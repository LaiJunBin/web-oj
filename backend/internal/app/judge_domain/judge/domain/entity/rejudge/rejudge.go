package rejudge

import "time"

type Rejudge struct {
	id               string
	contestProjectID int32
	reason           string
	status           *Status
	judgments        []string
	createdAt        time.Time
}

type RejudgeConfig struct {
	ID               string
	ContestProjectID int32
	Reason           string
	Status           *Status
	Judgments        []string
	CreatedAt        time.Time
}

func New(
	config *RejudgeConfig,
) *Rejudge {
	createdAt := time.Now()
	if !config.CreatedAt.IsZero() {
		createdAt = config.CreatedAt
	}

	return &Rejudge{
		id:               config.ID,
		contestProjectID: config.ContestProjectID,
		reason:           config.Reason,
		status:           config.Status,
		judgments:        config.Judgments,
		createdAt:        createdAt,
	}
}

func (r *Rejudge) ID() string {
	return r.id
}

// ContestProjectID is the contest project every judgment of this rejudge
// belongs to. Both entry points resolve to a single contest project: the
// whole-project rejudge is given one, and a single-submission rejudge takes the
// submission's. It is carried on the rejudge itself rather than derived from
// the judgments, because a project with no submissions still produces a
// rejudge — one with no judgments to derive anything from.
func (r *Rejudge) ContestProjectID() int32 {
	return r.contestProjectID
}

func (r *Rejudge) Reason() string {
	return r.reason
}

func (r *Rejudge) Status() *Status {
	return r.status
}

func (r *Rejudge) SetStatus(status *Status) {
	r.status = status
}

func (r *Rejudge) Judgments() []string {
	return r.judgments
}

func (r *Rejudge) AddJudgmentID(judgmentID string) {
	r.judgments = append(r.judgments, judgmentID)
}

func (r *Rejudge) CreatedAt() time.Time {
	return r.createdAt
}
