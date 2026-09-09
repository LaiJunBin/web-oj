package repository

import (
	"fmt"

	"judge-backend/internal/app/contest_domain/contest/domain/entity/contest"
	judgeprioritystrategy "judge-backend/internal/app/contest_domain/contest/domain/entity/contest/judge_priority_strategy"
	domainprovider "judge-backend/internal/app/contest_domain/contest/domain/provider"
	"judge-backend/internal/app/contest_domain/contest/domain/repository"
	"judge-backend/internal/app/shared/model"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/shared/types/httperror"

	"gorm.io/gorm"
)

type ContestRepository struct {
	db                *gorm.DB
	strategyValidator domainprovider.JudgePriorityStrategyValidator
	logger            sharedtypes.Logger
}

func NewContestRepository(
	db *gorm.DB,
	strategyValidator domainprovider.JudgePriorityStrategyValidator,
	logger sharedtypes.Logger,
) repository.ContestRepository {
	return &ContestRepository{
		db:                db,
		strategyValidator: strategyValidator,
		logger:            logger,
	}
}

func (r *ContestRepository) toDomainModel(c *model.Contest) (*contest.Contest, error) {
	config := &contest.ContestConfig{
		Name:             c.Name,
		Description:      c.Description,
		ActivationTime:   c.ActivateTime,
		DeactivationTime: c.DeactivateTime,
		IsPublic:         c.IsPublic,
		Projects: func() []contest.Project {
			projects := make([]contest.Project, len(c.Projects))
			for i, p := range c.Projects {
				var jps judgeprioritystrategy.JudgePriorityStrategy
				if r.strategyValidator.HasStrategy(p.JudgePriorityStrategy) {
					jps = judgeprioritystrategy.NewFromValidated(p.JudgePriorityStrategy)
				} else {
					// A contest silently running the WRONG scoring strategy
					// is a data problem, not a style one — make it visible.
					r.logger.Warnf("[ContestRepository] invalid judge priority strategy %q for contest project %d, using default", p.JudgePriorityStrategy, p.ID)
					jps = judgeprioritystrategy.NewDefault()
				}

				projects[i], _ = contest.NewProject(&contest.ProjectConfig{
					ID:                        p.ID,
					ProjectID:                 p.ProjectID,
					Name:                      p.Name,
					ActivationTime:            p.ActivateTime,
					DeactivationTime:          p.DeactivateTime,
					StartTime:                 p.StartTime,
					EndTime:                   p.EndTime,
					IsPublic:                  p.IsPublic,
					JudgePriorityStrategy:     jps,
					MaxJudgeDurationMinutes:   p.MaxJudgeDurationMinutes,
					SubmissionCooldownSeconds: p.SubmissionCooldownSeconds,
					PidsLimit:                 p.PidsLimit,
					MemoryLimitMB:             p.MemoryLimitMB,
					ScoreboardCompact:         p.ScoreboardCompact,
					ScoreboardMaskNames:       p.ScoreboardMaskNames,
				})
			}
			return projects
		}(),
	}

	contest, err := contest.New(config)
	if err != nil {
		return nil, err
	}

	contest.SetID(c.ID)
	contest.SetCreatedAt(c.CreatedAt)
	contest.SetUpdatedAt(c.UpdatedAt)

	return contest, nil
}

func (r *ContestRepository) Save(c *contest.Contest) error {
	projects := make([]model.ContestProject, 0)

	for _, p := range c.Projects() {
		if r.db.First(&model.Project{}, "id = ?", p.ProjectID()).Error != nil {
			return httperror.New(400, fmt.Sprintf("project with id %d not found", p.ProjectID()))
		}
		projects = append(projects, model.ContestProject{
			ID:                        p.ID(),
			ContestID:                 c.ID(),
			ProjectID:                 p.ProjectID(),
			Name:                      p.Name(),
			ActivateTime:              p.ActivationPeriod().StartTime(),
			DeactivateTime:            p.ActivationPeriod().EndTime(),
			StartTime:                 p.SubmissionPeriod().StartTime(),
			EndTime:                   p.SubmissionPeriod().EndTime(),
			IsPublic:                  p.IsPublic(),
			JudgePriorityStrategy:     p.JudgePriorityStrategy().String(),
			MaxJudgeDurationMinutes:   p.MaxJudgeDurationMinutes(),
			SubmissionCooldownSeconds: p.SubmissionCooldownSeconds(),
			PidsLimit:                 p.PidsLimit(),
			MemoryLimitMB:             p.MemoryLimitMB(),
			ScoreboardCompact:         p.ScoreboardCompact(),
			ScoreboardMaskNames:       p.ScoreboardMaskNames(),
		})
	}

	contest := &model.Contest{
		ID:             c.ID(),
		Name:           c.Name().String(),
		Description:    c.Description(),
		ActivateTime:   c.ActivationPeriod().StartTime(),
		DeactivateTime: c.ActivationPeriod().EndTime(),
		IsPublic:       c.IsPublic(),
		Projects:       projects,
	}

	contest.CreatedAt = c.CreatedAt()

	tx := r.db.Begin()

	if err := tx.Omit("Projects").Save(contest).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	c.SetID(contest.ID)

	return nil
}

func (r *ContestRepository) FindByID(id int32) (*contest.Contest, error) {
	cModel := &model.Contest{}
	if err := r.db.Preload("Projects").First(cModel, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, httperror.New(404, "contest not found")
		}
		return nil, err
	}
	return r.toDomainModel(cModel)
}

func (r *ContestRepository) Delete(c *contest.Contest) error {
	if err := r.db.Delete(&model.Contest{}, "id = ?", c.ID()).Error; err != nil {
		return err
	}
	return nil
}
