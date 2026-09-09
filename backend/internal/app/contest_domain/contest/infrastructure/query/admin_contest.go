package query

import (
	"judge-backend/internal/app/contest_domain/contest/application/query"
	"judge-backend/internal/app/contest_domain/contest/application/query/dto"
	"judge-backend/internal/app/shared/model"
	"judge-backend/internal/app/shared/types/httperror"
	"net/http"

	"gorm.io/gorm"
)

type AdminContestQuery struct {
	db *gorm.DB
}

func NewAdminContestQuery(
	db *gorm.DB,
) query.AdminContestQuery {
	return &ContestQuery{
		db: db,
	}
}

func (q *ContestQuery) ListContest() ([]*dto.AdminContest, error) {
	contests := []*model.Contest{}
	if err := q.db.Preload("Projects").Find(&contests).Error; err != nil {
		return nil, httperror.From(err)
	}

	result := make([]*dto.AdminContest, 0, len(contests))
	for _, c := range contests {
		result = append(result, &dto.AdminContest{
			ID:               c.ID,
			Name:             c.Name,
			Description:      c.Description,
			IsPublic:         c.IsPublic,
			ActivationTime:   c.ActivateTime,
			DeactivationTime: c.DeactivateTime,
			TotalProjects:    int32(len(c.Projects)),
			CreatedAt:        c.CreatedAt,
			UpdatedAt:        c.UpdatedAt,
		})
	}

	return result, nil
}

func (q *ContestQuery) GetContestByID(id int32) (*dto.AdminContestDetail, error) {
	contest := &model.Contest{}
	if err := q.db.Preload("Projects", func(tx *gorm.DB) *gorm.DB {
		return tx.Preload("Project")
	}).Where("id = ?", id).First(contest).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, httperror.New(http.StatusNotFound, "contest not found")
		}

		return nil, httperror.From(err)
	}

	projects := make([]*dto.AdminContestProject, 0, len(contest.Projects))
	for _, p := range contest.Projects {
		projects = append(projects, &dto.AdminContestProject{
			ID:                        p.ID,
			ProjectID:                 p.ProjectID,
			Name:                      p.Name,
			Image:                     p.Project.ImageTag,
			TestImage:                 p.Project.TestImageTag,
			Description:               p.Project.Description,
			IsPublic:                  p.IsPublic,
			ActivateTime:              p.ActivateTime,
			DeactivateTime:            p.DeactivateTime,
			StartTime:                 p.StartTime,
			EndTime:                   p.EndTime,
			JudgePriorityStrategy:     p.JudgePriorityStrategy,
			MaxJudgeDurationMinutes:   p.MaxJudgeDurationMinutes,
			SubmissionCooldownSeconds: p.SubmissionCooldownSeconds,
			PidsLimit:                 p.PidsLimit,
			MemoryLimitMB:             p.MemoryLimitMB,
			ScoreboardCompact:         p.ScoreboardCompact,
			ScoreboardMaskNames:       p.ScoreboardMaskNames,
		})
	}

	return &dto.AdminContestDetail{
		ID:               contest.ID,
		Name:             contest.Name,
		Description:      contest.Description,
		IsPublic:         contest.IsPublic,
		Projects:         projects,
		ActivationTime:   contest.ActivateTime,
		DeactivationTime: contest.DeactivateTime,
		CreatedAt:        contest.CreatedAt,
		UpdatedAt:        contest.UpdatedAt,
	}, nil
}
