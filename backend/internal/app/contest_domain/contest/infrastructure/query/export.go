package query

import (
	"errors"

	"judge-backend/internal/app/contest_domain/contest/application/query"
	"judge-backend/internal/app/contest_domain/contest/application/query/dto"
	"judge-backend/internal/app/shared/model"

	"gorm.io/gorm"
)

type ContestExportQuery struct {
	db *gorm.DB
}

func NewContestExportQuery(db *gorm.DB) query.ContestExportQuery {
	return &ContestExportQuery{db: db}
}

// LoadContestForExport reads the contest and its slots.
//
// Slots are ordered explicitly. Without an ORDER BY the database may return
// them however it likes, and two exports of an unchanged contest would differ
// — which defeats a bundle meant to be diffed in review.
func (q ContestExportQuery) LoadContestForExport(id int32) (*dto.ContestExport, error) {
	m := model.Contest{}
	err := q.db.
		Preload("Projects", func(db *gorm.DB) *gorm.DB { return db.Order("id ASC") }).
		Preload("Projects.Project").
		Where("id = ?", id).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, query.ErrorContestNotFound
	}
	if err != nil {
		return nil, err
	}

	out := &dto.ContestExport{
		ID:             m.ID,
		Name:           m.Name,
		Description:    m.Description,
		IsPublic:       m.IsPublic,
		ActivateTime:   m.ActivateTime,
		DeactivateTime: m.DeactivateTime,
	}

	for _, p := range m.Projects {
		out.Projects = append(out.Projects, dto.ContestProjectExport{
			ProjectName:               p.Project.Name,
			Name:                      p.Name,
			IsPublic:                  p.IsPublic,
			JudgePriorityStrategy:     p.JudgePriorityStrategy,
			ActivateTime:              p.ActivateTime,
			DeactivateTime:            p.DeactivateTime,
			StartTime:                 p.StartTime,
			EndTime:                   p.EndTime,
			MaxJudgeDurationMinutes:   p.MaxJudgeDurationMinutes,
			SubmissionCooldownSeconds: p.SubmissionCooldownSeconds,
			PidsLimit:                 p.PidsLimit,
			MemoryLimitMB:             p.MemoryLimitMB,
			ScoreboardCompact:         p.ScoreboardCompact,
			ScoreboardMaskNames:       p.ScoreboardMaskNames,
		})
	}

	return out, nil
}
