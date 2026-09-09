package updatecontest

import (
	"judge-backend/internal/app/contest_domain/contest/domain/entity/contest"
	judgeprioritystrategy "judge-backend/internal/app/contest_domain/contest/domain/entity/contest/judge_priority_strategy"
	domainprovider "judge-backend/internal/app/contest_domain/contest/domain/provider"
	"judge-backend/internal/app/contest_domain/contest/domain/repository"
	"judge-backend/internal/app/shared/types/httperror"
	"net/http"
	"time"
)

var (
	ErrorInvalidID                = httperror.New(http.StatusBadRequest, "invalid id, must be greater than 0")
	ErrorNoInput                  = httperror.New(http.StatusBadRequest, "no input")
	ErrorNameIsRequired           = httperror.New(http.StatusBadRequest, "name is required")
	ErrorActivateTimeIsRequired   = httperror.New(http.StatusBadRequest, "activateTime is required")
	ErrorDeactivateTimeIsRequired = httperror.New(http.StatusBadRequest, "deactivateTime is required")
)

type UpdateProjectConfig struct {
	ID                    *int32
	ProjectID             *int32
	Name                  *string
	ActivationTime        *time.Time
	DeactivationTime      *time.Time
	StartTime             *time.Time
	EndTime               *time.Time
	IsPublic              *bool
	JudgePriorityStrategy *string
	// MaxJudgeDurationMinutes: absent = keep current value; the API cannot
	// clear it back to "use global default" today (acceptable — set a value
	// equal to the global default instead).
	MaxJudgeDurationMinutes *int32
	// SubmissionCooldownSeconds: same partial-update semantics as
	// MaxJudgeDurationMinutes — absent keeps the current value. Sending 0 is
	// NOT "clear it": it explicitly turns the cooldown off for this project.
	SubmissionCooldownSeconds *int32
	// PidsLimit and MemoryLimitMB: same again — absent keeps the current
	// value, and 0 is an explicit "no cap for this project", not a clear.
	PidsLimit           *int32
	MemoryLimitMB       *int32
	ScoreboardCompact   *bool
	ScoreboardMaskNames *bool
}

type Input struct {
	ID             int32
	Name           *string
	Description    *string
	ActivateTime   *time.Time
	DeactivateTime *time.Time
	IsPublic       *bool
	ProjectConfigs *[]UpdateProjectConfig
}

type Output struct {
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	strategyValidator domainprovider.JudgePriorityStrategyValidator
	repository        repository.ContestRepository
}

func NewUsecase(
	strategyValidator domainprovider.JudgePriorityStrategyValidator,
	repository repository.ContestRepository,
) Usecase {
	return &usecase{
		strategyValidator: strategyValidator,
		repository:        repository,
	}
}

func (u *usecase) validate(input *Input) error {
	if input.ID <= 0 {
		return ErrorInvalidID
	}

	if input.Name == nil &&
		input.Description == nil &&
		input.ActivateTime == nil &&
		input.DeactivateTime == nil &&
		input.IsPublic == nil &&
		input.ProjectConfigs == nil {
		return ErrorNoInput
	}

	if input.Name != nil && *input.Name == "" {
		return ErrorNameIsRequired
	}

	if input.ActivateTime != nil && input.ActivateTime.IsZero() {
		return ErrorActivateTimeIsRequired
	}

	if input.DeactivateTime != nil && input.DeactivateTime.IsZero() {
		return ErrorDeactivateTimeIsRequired
	}

	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	if err := u.validate(input); err != nil {
		return nil, err
	}

	c, err := u.repository.FindByID(input.ID)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		if err := c.SetName(*input.Name); err != nil {
			return nil, err
		}
	}

	if input.Description != nil {
		c.SetDescription(*input.Description)
	}

	if input.IsPublic != nil {
		c.SetIsPublic(*input.IsPublic)
	}

	if input.ActivateTime != nil || input.DeactivateTime != nil {
		times := []time.Time{c.ActivationPeriod().StartTime(), c.ActivationPeriod().EndTime()}
		if input.ActivateTime != nil {
			times[0] = *input.ActivateTime
		}

		if input.DeactivateTime != nil {
			times[1] = *input.DeactivateTime
		}

		if err := c.SetActivationPeriod(times[0], times[1]); err != nil {
			return nil, err
		}
	}

	if input.ProjectConfigs != nil {
		projectConfigs := make([]contest.ProjectConfig, len(c.Projects()))
		projectConfigMap := make(map[int32]contest.ProjectConfig)
		for i, p := range c.Projects() {
			projectConfigs[i] = contest.ProjectConfig{
				ID:                        p.ID(),
				Name:                      p.Name(),
				ProjectID:                 p.ProjectID(),
				ActivationTime:            p.ActivationPeriod().StartTime(),
				DeactivationTime:          p.ActivationPeriod().EndTime(),
				StartTime:                 p.SubmissionPeriod().StartTime(),
				EndTime:                   p.SubmissionPeriod().EndTime(),
				IsPublic:                  p.IsPublic(),
				JudgePriorityStrategy:     p.JudgePriorityStrategy(),
				MaxJudgeDurationMinutes:   p.MaxJudgeDurationMinutes(),
				SubmissionCooldownSeconds: p.SubmissionCooldownSeconds(),
				PidsLimit:                 p.PidsLimit(),
				MemoryLimitMB:             p.MemoryLimitMB(),
				ScoreboardCompact:         p.ScoreboardCompact(),
				ScoreboardMaskNames:       p.ScoreboardMaskNames(),
			}
			projectConfigMap[p.ID()] = projectConfigs[i]
		}

		configs := make([]contest.ProjectConfig, 0)
		for _, p := range *input.ProjectConfigs {
			config := contest.ProjectConfig{}
			if p.ID != nil {
				config = projectConfigMap[*p.ID]
			}

			if p.ProjectID != nil {
				config.ProjectID = *p.ProjectID
			}

			if p.Name != nil {
				config.Name = *p.Name
			}

			if p.ActivationTime != nil {
				config.ActivationTime = *p.ActivationTime
			}

			if p.DeactivationTime != nil {
				config.DeactivationTime = *p.DeactivationTime
			}

			if p.StartTime != nil {
				config.StartTime = *p.StartTime
			}

			if p.EndTime != nil {
				config.EndTime = *p.EndTime
			}

			if p.IsPublic != nil {
				config.IsPublic = *p.IsPublic
			}

			if p.JudgePriorityStrategy != nil {
				var jps judgeprioritystrategy.JudgePriorityStrategy
				if u.strategyValidator.HasStrategy(*p.JudgePriorityStrategy) {
					jps = judgeprioritystrategy.NewFromValidated(*p.JudgePriorityStrategy)
				} else {
					jps = judgeprioritystrategy.NewDefault()
				}

				config.JudgePriorityStrategy = jps
			}

			if p.MaxJudgeDurationMinutes != nil {
				config.MaxJudgeDurationMinutes = p.MaxJudgeDurationMinutes
			}

			if p.SubmissionCooldownSeconds != nil {
				config.SubmissionCooldownSeconds = p.SubmissionCooldownSeconds
			}

			if p.PidsLimit != nil {
				config.PidsLimit = p.PidsLimit
			}

			if p.MemoryLimitMB != nil {
				config.MemoryLimitMB = p.MemoryLimitMB
			}

			if p.ScoreboardCompact != nil {
				config.ScoreboardCompact = *p.ScoreboardCompact
			}

			if p.ScoreboardMaskNames != nil {
				config.ScoreboardMaskNames = *p.ScoreboardMaskNames
			}

			configs = append(configs, config)
		}

		projects, err := contest.NewProjects(configs)
		if err != nil {
			return nil, err
		}

		c.SetProjects(projects)
	}

	if err := u.repository.Save(c); err != nil {
		return nil, err
	}

	return nil, nil
}
