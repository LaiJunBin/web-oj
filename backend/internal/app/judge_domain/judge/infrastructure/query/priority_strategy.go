package query

import (
	"judge-backend/internal/app/judge_domain/judge/domain/query"
	"judge-backend/internal/app/shared/model"
	sharedtypes "judge-backend/internal/app/shared/types"

	"gorm.io/gorm"
)

type PriorityStrategyQuery struct {
	db     *gorm.DB
	logger sharedtypes.Logger
}

func NewPriorityStrategyQuery(
	db *gorm.DB,
	logger sharedtypes.Logger,
) query.PriorityStrategyQuery {
	logger.Infof("[PriorityStrategyQuery] Initializing new priority strategy query")
	return &PriorityStrategyQuery{
		db:     db,
		logger: logger,
	}
}

func (r *PriorityStrategyQuery) GetStrategyNameByContestProjectID(contestProjectID int32) (string, error) {
	r.logger.Infof("[PriorityStrategyQuery] Getting strategy name for contest project ID: %d", contestProjectID)
	cp := &model.ContestProject{}
	if err := r.db.Where("id = ?", contestProjectID).First(cp).Error; err != nil {
		r.logger.Errorf("[PriorityStrategyQuery] Failed to get strategy name for contest project ID %d: %v", contestProjectID, err)
		return "", err
	}

	r.logger.Infof("[PriorityStrategyQuery] Successfully retrieved strategy name: %s for contest project ID: %d", cp.JudgePriorityStrategy, contestProjectID)
	return cp.JudgePriorityStrategy, nil
}
