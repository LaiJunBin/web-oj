package query

import (
	"judge-backend/internal/app/project_domain/project/application/query"
	"judge-backend/internal/app/project_domain/project/application/query/dto"
	"judge-backend/internal/app/shared/model"
	"judge-backend/internal/app/shared/types/httperror"
	"net/http"

	"gorm.io/gorm"
)

type TestcaseQuery struct {
	db *gorm.DB
}

func NewTestcaseQuery(db *gorm.DB) query.TestcaseQuery {
	return &TestcaseQuery{
		db: db,
	}
}

func (q *TestcaseQuery) ListProjectTestcases(projectID int32) ([]*dto.ProjectTestcaseGradingItem, error) {
	project := &model.Project{}
	if err := q.db.Preload("GradingItems", func(db *gorm.DB) *gorm.DB {
		return db.Preload("Testcases", func(db *gorm.DB) *gorm.DB {
			return db.Select("grading_item_id, id, name")
		}).Select("project_id, id, name, label")
	}).First(project, "id = ?", projectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, httperror.New(http.StatusNotFound, "project not found")
		}
		return nil, err
	}

	items := make([]*dto.ProjectTestcaseGradingItem, 0, len(project.GradingItems))
	for _, gi := range project.GradingItems {
		testcases := make([]*dto.ProjectTestcaseInfo, 0, len(gi.Testcases))
		for _, tc := range gi.Testcases {
			testcases = append(testcases, &dto.ProjectTestcaseInfo{
				ID:   tc.ID,
				Name: tc.Name,
			})
		}

		items = append(items, &dto.ProjectTestcaseGradingItem{
			ID:        gi.ID,
			Label:     gi.Label,
			Name:      gi.Name,
			Testcases: testcases,
		})
	}

	return items, nil
}

func (q *TestcaseQuery) GetTestcase(id int32) (*dto.TestcaseDetail, error) {
	tc := &model.Testcase{}
	if err := q.db.Preload("Depends", func(db *gorm.DB) *gorm.DB {
		return db.Preload("GradingItem", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, name, label")
		})
	}).First(tc, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, err
	}

	t := &dto.TestcaseDetail{
		Testcase: dto.Testcase{
			TestcaseInfo: dto.TestcaseInfo{
				ID:        tc.ID,
				Name:      tc.Name,
				Score:     tc.Score,
				CreatedAt: tc.CreatedAt,
				UpdatedAt: tc.UpdatedAt,
			},
			IsSample: tc.IsSample,
			Command:  tc.Command,
		},
		Dependencies: make([]*dto.DependencyTestcase, 0, len(tc.Depends)),
	}

	for _, d := range tc.Depends {
		t.Dependencies = append(t.Dependencies, &dto.DependencyTestcase{
			Testcase: dto.DependencyTestcaseInfo{
				ID:   d.ID,
				Name: d.Name,
			},
			GradingItem: dto.GradingItemInfo{
				ID:    d.GradingItemID,
				Label: d.GradingItem.Label,
				Name:  d.GradingItem.Name,
			},
		})
	}

	return t, nil
}
