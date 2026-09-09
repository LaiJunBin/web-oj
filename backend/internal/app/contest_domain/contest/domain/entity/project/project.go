package project

import (
	"judge-backend/internal/app/shared/types/httperror"
	"net/http"
	"time"
)

type Project struct {
	id               int32
	name             Name
	isPublic         bool
	activationPeriod DurationPeriod
	submissionPeriod DurationPeriod
}

type ProjectConfig struct {
	ID             int32
	Name           string
	IsPublic       bool
	ActivateTime   time.Time
	DeactivateTime time.Time
	StartTime      time.Time
	EndTime        time.Time
}

func New(config *ProjectConfig) (*Project, error) {
	name, err := NewName(config.Name)
	if err != nil {
		return nil, httperror.New(http.StatusBadRequest, "invalid project name")
	}

	activationPeriod, err := NewDurationPeriod(config.ActivateTime, config.DeactivateTime)
	if err != nil {
		return nil, httperror.New(http.StatusBadRequest, "invalid activation period")
	}

	submissionPeriod, err := NewDurationPeriod(config.StartTime, config.EndTime)
	if err != nil {
		return nil, httperror.New(http.StatusBadRequest, "invalid submission period")
	}

	return &Project{
		id:               config.ID,
		name:             name,
		isPublic:         config.IsPublic,
		activationPeriod: activationPeriod,
		submissionPeriod: submissionPeriod,
	}, nil
}

func (p *Project) ID() int32 {
	return p.id
}

func (p *Project) Name() Name {
	return p.name
}

func (p *Project) IsPublic() bool {
	return p.isPublic
}

func (p *Project) ActivationPeriod() DurationPeriod {
	return p.activationPeriod
}

func (p *Project) SubmissionPeriod() DurationPeriod {
	return p.submissionPeriod
}

func (p *Project) IsWithinActivationPeriod(currentTime time.Time) bool {
	return p.activationPeriod.IsWithin(currentTime)
}

func (p *Project) IsWithinSubmissionPeriod(currentTime time.Time) bool {
	return p.submissionPeriod.IsWithin(currentTime)
}

// IsSubmissionNotStarted mirrors the read side's model.ContestProject.IsStarted
// (StartTime.Before(now)): exactly at StartTime the window is not yet open.
// The submit gate uses this pair instead of IsWithinSubmissionPeriod because
// the latter treats the EndTime instant as outside the window, which would
// disagree with what IsEnded shows contestants.
func (p *Project) IsSubmissionNotStarted(currentTime time.Time) bool {
	return !p.submissionPeriod.StartTime().Before(currentTime)
}

// IsSubmissionEnded mirrors model.ContestProject.IsEnded (EndTime.Before(now)):
// exactly at EndTime the window is still open.
func (p *Project) IsSubmissionEnded(currentTime time.Time) bool {
	return p.submissionPeriod.EndTime().Before(currentTime)
}

func (p *Project) TimeLeftToStartSubmit(currentTime time.Time) time.Duration {
	return p.submissionPeriod.TimeLeftToStart(currentTime)
}

func (p *Project) TimeLeftToSubmit(currentTime time.Time) time.Duration {
	return p.submissionPeriod.TimeLeftToEnd(currentTime)
}
