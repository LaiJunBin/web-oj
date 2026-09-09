package query

import "judge-backend/internal/app/project_domain/project/application/query/dto"

type GradingItemQuery interface {
	ListGradingItem(projectID int32) ([]*dto.GradingItem, error)
	GetGradingItem(id int32) (*dto.GradingItemDetail, error)
}
