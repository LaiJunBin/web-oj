package contest

import (
	"fmt"
	judgeprioritystrategy "judge-backend/internal/app/contest_domain/contest/domain/entity/contest/judge_priority_strategy"
	"judge-backend/internal/app/shared/types/httperror"
	"net/http"
	"time"
)

type Project struct {
	id                    int32
	projectID             int32
	name                  string
	activationPeriod      ActivationPeriod
	submissionPeriod      SubmissionPeriod
	isPublic              bool
	judgePriorityStrategy judgeprioritystrategy.JudgePriorityStrategy
	// maxJudgeDurationMinutes caps one judgment's total wall-clock for this
	// contest project; nil means "use the global judge config default".
	maxJudgeDurationMinutes   *int32
	submissionCooldownSeconds *int32
	// pidsLimit and memoryLimitMB cap a judging container's processes and
	// memory. nil means "use the global judge config default"; 0 means no cap.
	pidsLimit           *int32
	memoryLimitMB       *int32
	scoreboardCompact   bool
	scoreboardMaskNames bool
}

type ProjectConfig struct {
	ID                        int32
	ProjectID                 int32
	Name                      string
	ActivationTime            time.Time
	DeactivationTime          time.Time
	StartTime                 time.Time
	EndTime                   time.Time
	IsPublic                  bool
	JudgePriorityStrategy     judgeprioritystrategy.JudgePriorityStrategy
	MaxJudgeDurationMinutes   *int32
	SubmissionCooldownSeconds *int32
	PidsLimit                 *int32
	MemoryLimitMB             *int32
	ScoreboardCompact         bool
	ScoreboardMaskNames       bool
}

func NewProject(config *ProjectConfig) (Project, error) {
	if config.ID < 0 {
		return Project{}, httperror.New(http.StatusBadRequest, "invalid id")
	}

	if config.ProjectID <= 0 {
		return Project{}, httperror.New(http.StatusBadRequest, "invalid project id")
	}

	activePeriod, err := NewActivationPeriod(config.ActivationTime, config.DeactivationTime)
	if err != nil {
		return Project{}, httperror.New(http.StatusBadRequest, "invalid activation period")
	}

	submissionPeriod, err := NewSubmissionPeriod(config.StartTime, config.EndTime)
	if err != nil {
		return Project{}, httperror.New(http.StatusBadRequest, "invalid submission period")
	}

	if !config.JudgePriorityStrategy.Valid() {
		return Project{}, httperror.New(http.StatusBadRequest, "invalid judge priority strategy")
	}

	if config.MaxJudgeDurationMinutes != nil && *config.MaxJudgeDurationMinutes <= 0 {
		return Project{}, httperror.New(http.StatusBadRequest, "maxJudgeDurationMinutes must be positive")
	}

	// Zero is allowed, unlike maxJudgeDurationMinutes: it is how a contest
	// says "no cooldown here" and overrides a non-zero system default. Only
	// a negative wait is meaningless.
	if config.SubmissionCooldownSeconds != nil && *config.SubmissionCooldownSeconds < 0 {
		return Project{}, httperror.New(http.StatusBadRequest, "submissionCooldownSeconds must not be negative")
	}

	// Same reasoning as submissionCooldownSeconds: 0 is the contest project's
	// way to say "no cap here" and override a non-zero deployment default,
	// which is the whole reason these columns are nullable. Only a negative
	// cap is meaningless.
	if config.PidsLimit != nil && *config.PidsLimit < 0 {
		return Project{}, httperror.New(http.StatusBadRequest, "pidsLimit must not be negative")
	}

	if config.MemoryLimitMB != nil && *config.MemoryLimitMB < 0 {
		return Project{}, httperror.New(http.StatusBadRequest, "memoryLimitMB must not be negative")
	}

	return Project{
		id:                        config.ID,
		projectID:                 config.ProjectID,
		name:                      config.Name,
		activationPeriod:          activePeriod,
		submissionPeriod:          submissionPeriod,
		isPublic:                  config.IsPublic,
		judgePriorityStrategy:     config.JudgePriorityStrategy,
		maxJudgeDurationMinutes:   config.MaxJudgeDurationMinutes,
		submissionCooldownSeconds: config.SubmissionCooldownSeconds,
		pidsLimit:                 config.PidsLimit,
		memoryLimitMB:             config.MemoryLimitMB,
		scoreboardCompact:         config.ScoreboardCompact,
		scoreboardMaskNames:       config.ScoreboardMaskNames,
	}, nil
}

func NewProjects(configs []ProjectConfig) ([]Project, error) {
	projects := make([]Project, len(configs))
	for i, config := range configs {
		project, err := NewProject(&config)
		if err != nil {
			return nil, httperror.From(err).WithDetails(fmt.Sprintf("project index: %d", i))
		}

		projects[i] = project
	}

	return projects, nil
}

func (p Project) ID() int32 {
	return p.id
}

func (p Project) ProjectID() int32 {
	return p.projectID
}

func (p Project) Name() string {
	return p.name
}

func (p Project) ActivationPeriod() ActivationPeriod {
	return p.activationPeriod
}

func (p Project) SubmissionPeriod() SubmissionPeriod {
	return p.submissionPeriod
}

func (p Project) IsPublic() bool {
	return p.isPublic
}

func (p Project) JudgePriorityStrategy() judgeprioritystrategy.JudgePriorityStrategy {
	return p.judgePriorityStrategy
}

func (p Project) MaxJudgeDurationMinutes() *int32 {
	return p.maxJudgeDurationMinutes
}

func (p Project) SubmissionCooldownSeconds() *int32 {
	return p.submissionCooldownSeconds
}

func (p Project) PidsLimit() *int32 {
	return p.pidsLimit
}

func (p Project) MemoryLimitMB() *int32 {
	return p.memoryLimitMB
}

func (p Project) ScoreboardCompact() bool {
	return p.scoreboardCompact
}

func (p Project) ScoreboardMaskNames() bool {
	return p.scoreboardMaskNames
}

func (p *Project) SetID(id int32) {
	p.id = id
}
