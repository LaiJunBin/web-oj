package query

import "judge-backend/internal/app/project_domain/project/application/query/dto"

type TestcaseQuery interface {
	ListProjectTestcases(projectID int32) ([]*dto.ProjectTestcaseGradingItem, error)
	GetTestcase(id int32) (*dto.TestcaseDetail, error)
}
