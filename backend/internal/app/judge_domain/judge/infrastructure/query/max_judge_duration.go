package query

import (
	"judge-backend/internal/app/judge_domain/judge/domain/query"
	"judge-backend/internal/app/shared/model"
	sharedtypes "judge-backend/internal/app/shared/types"

	"gorm.io/gorm"
)

type MaxJudgeDurationQuery struct {
	db     *gorm.DB
	logger sharedtypes.Logger
}

func NewMaxJudgeDurationQuery(
	db *gorm.DB,
	logger sharedtypes.Logger,
) query.MaxJudgeDurationQuery {
	return &MaxJudgeDurationQuery{
		db:     db,
		logger: logger,
	}
}

func (r *MaxJudgeDurationQuery) GetMaxJudgeDurationMinutesByContestProjectID(contestProjectID int32) (*int32, error) {
	cp := &model.ContestProject{}
	if err := r.db.Select("max_judge_duration_minutes").Where("id = ?", contestProjectID).First(cp).Error; err != nil {
		r.logger.Errorf("[MaxJudgeDurationQuery] Failed to get max judge duration for contest project %d: %v", contestProjectID, err)
		return nil, err
	}
	return cp.MaxJudgeDurationMinutes, nil
}
