package createcontest

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
	ErrorNameIsRequired           = httperror.New(http.StatusBadRequest, "name is required")
	ErrorActivateTimeIsRequired   = httperror.New(http.StatusBadRequest, "activateTime is required")
	ErrorDeactivateTimeIsRequired = httperror.New(http.StatusBadRequest, "deactivateTime is required")
)

type CreateProjectConfig struct {
	ProjectID                 int32
	Name                      string
	ActivationTime            time.Time
	DeactivationTime          time.Time
	StartTime                 time.Time
	EndTime                   time.Time
	IsPublic                  bool
	JudgePriorityStrategy     string
	MaxJudgeDurationMinutes   *int32
	SubmissionCooldownSeconds *int32
	PidsLimit                 *int32
	MemoryLimitMB             *int32
	ScoreboardCompact         bool
	ScoreboardMaskNames       bool
}

type Input struct {
	Name           string
	Description    string
	ActivateTime   time.Time
	DeactivateTime time.Time
	IsPublic       bool
	ProjectConfigs []CreateProjectConfig
}

type Output struct {
	ID int32
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
	if input.Name == "" {
		return ErrorNameIsRequired
	}

	if input.ActivateTime.IsZero() {
		return ErrorActivateTimeIsRequired
	}

	if input.DeactivateTime.IsZero() {
		return ErrorDeactivateTimeIsRequired
	}

	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	if err := u.validate(input); err != nil {
		return nil, err
	}

	configs := make([]contest.ProjectConfig, 0)
	for _, project := range input.ProjectConfigs {
		var jps judgeprioritystrategy.JudgePriorityStrategy
		if u.strategyValidator.HasStrategy(project.JudgePriorityStrategy) {
			jps = judgeprioritystrategy.NewFromValidated(project.JudgePriorityStrategy)
		} else {
			jps = judgeprioritystrategy.NewDefault()
		}

		configs = append(configs, contest.ProjectConfig{
			ProjectID:                 project.ProjectID,
			Name:                      project.Name,
			ActivationTime:            project.ActivationTime,
			DeactivationTime:          project.DeactivationTime,
			StartTime:                 project.StartTime,
			EndTime:                   project.EndTime,
			IsPublic:                  project.IsPublic,
			JudgePriorityStrategy:     jps,
			MaxJudgeDurationMinutes:   project.MaxJudgeDurationMinutes,
			SubmissionCooldownSeconds: project.SubmissionCooldownSeconds,
			PidsLimit:                 project.PidsLimit,
			MemoryLimitMB:             project.MemoryLimitMB,
			ScoreboardCompact:         project.ScoreboardCompact,
			ScoreboardMaskNames:       project.ScoreboardMaskNames,
		})
	}

	projects, err := contest.NewProjects(configs)
	if err != nil {
		return nil, err
	}

	contest, err := contest.New(&contest.ContestConfig{
		Name:             input.Name,
		Description:      input.Description,
		ActivationTime:   input.ActivateTime,
		DeactivationTime: input.DeactivateTime,
		IsPublic:         input.IsPublic,
		Projects:         projects,
	})

	if err != nil {
		return nil, err
	}

	if err := u.repository.Save(contest); err != nil {
		return nil, err
	}

	return &Output{
		ID: contest.ID(),
	}, nil
}
