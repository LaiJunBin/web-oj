package contest

import (
	judgeprioritystrategy "judge-backend/internal/app/contest_domain/contest/domain/entity/contest/judge_priority_strategy"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/shared/types/httperror"
	"net/http"
	"time"
)

type Contest struct {
	id                    int32
	name                  Name
	description           string
	activationPeriod      ActivationPeriod
	isPublic              bool
	judgePriorityStrategy *judgeprioritystrategy.JudgePriorityStrategy

	projects []Project

	sharedtypes.TimeStamps
}

type ContestConfig struct {
	Name                  string
	Description           string
	ActivationTime        time.Time
	DeactivationTime      time.Time
	IsPublic              bool
	Projects              []Project
	JudgePriorityStrategy *judgeprioritystrategy.JudgePriorityStrategy
}

func New(config *ContestConfig) (*Contest, error) {
	contest := &Contest{
		description:           config.Description,
		projects:              config.Projects,
		isPublic:              config.IsPublic,
		judgePriorityStrategy: config.JudgePriorityStrategy,
	}

	if err := contest.SetName(config.Name); err != nil {
		return nil, err
	}

	if err := contest.SetActivationPeriod(config.ActivationTime, config.DeactivationTime); err != nil {
		return nil, err
	}

	return contest, nil
}

func (c *Contest) ID() int32 {
	return c.id
}

func (c *Contest) Name() Name {
	return c.name
}

func (c *Contest) Description() string {
	return c.description
}

func (c *Contest) ActivationPeriod() ActivationPeriod {
	return c.activationPeriod
}

func (c *Contest) IsActive(currentTime time.Time) bool {
	return c.activationPeriod.IsActive(currentTime)
}

func (c *Contest) IsPublic() bool {
	return c.isPublic
}

func (c *Contest) Projects() []Project {
	return c.projects
}

func (c *Contest) JudgePriorityStrategy() *judgeprioritystrategy.JudgePriorityStrategy {
	return c.judgePriorityStrategy
}

func (c *Contest) SetID(id int32) {
	c.id = id
}
func (c *Contest) SetName(name string) error {
	n, err := NewName(name)
	if err != nil {
		return httperror.New(http.StatusBadRequest, "invalid contest name")
	}

	c.name = n
	return nil
}

func (c *Contest) SetDescription(description string) {
	c.description = description
}

func (c *Contest) SetActivationPeriod(startTime, endTime time.Time) error {
	ap, err := NewActivationPeriod(startTime, endTime)
	if err != nil {
		return httperror.New(http.StatusBadRequest, "invalid activation period")
	}

	c.activationPeriod = ap
	return nil
}

func (c *Contest) SetIsPublic(isPublic bool) {
	c.isPublic = isPublic
}

func (c *Contest) SetProjects(projects []Project) {
	c.projects = projects
}

func (c *Contest) SetJudgePriorityStrategy(strategy *judgeprioritystrategy.JudgePriorityStrategy) {
	c.judgePriorityStrategy = strategy
}
