package query

import (
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/testcase/resultparser"
	"judge-backend/internal/app/project_domain/project/application/query"
	"judge-backend/internal/app/project_domain/project/application/query/dto"
	"judge-backend/internal/app/shared/model"

	"gorm.io/gorm"
)

type GradingItemQuery struct {
	db *gorm.DB
}

func NewGradingItemQuery(db *gorm.DB) query.GradingItemQuery {
	return &GradingItemQuery{
		db: db,
	}
}

func (q *GradingItemQuery) ListGradingItem(projectID int32) ([]*dto.GradingItem, error) {
	gradingItems := []*model.GradingItem{}
	if err := q.db.Preload("Testcases").Order("`order` ASC").Find(&gradingItems, "project_id = ?", projectID).Error; err != nil {
		return nil, err
	}

	result := make([]*dto.GradingItem, 0, len(gradingItems))
	for _, gradingItem := range gradingItems {
		result = append(result, &dto.GradingItem{
			ID:          gradingItem.ID,
			Name:        gradingItem.Name,
			Label:       gradingItem.Label,
			Description: gradingItem.Description,
			Order:       gradingItem.Order,
			Score:       gradingItem.Score,
			TestcaseLen: int32(len(gradingItem.Testcases)),
			CreatedAt:   gradingItem.CreatedAt,
			UpdatedAt:   gradingItem.UpdatedAt,
		})
	}

	return result, nil
}

func (q *GradingItemQuery) GetGradingItem(id int32) (*dto.GradingItemDetail, error) {
	gradingItem := &model.GradingItem{}
	if err := q.db.Preload("Testcases").Preload("Depends", func(db *gorm.DB) *gorm.DB {
		return db.Preload("GradingItem", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, name, label")
		})
	}).First(gradingItem, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, err
	}

	testcases := make([]*dto.Testcase, 0, len(gradingItem.Testcases))
	for _, testcase := range gradingItem.Testcases {
		testcases = append(testcases, &dto.Testcase{
			TestcaseInfo: dto.TestcaseInfo{
				ID:        testcase.ID,
				Name:      testcase.Name,
				Score:     testcase.Score,
				CreatedAt: testcase.CreatedAt,
				UpdatedAt: testcase.UpdatedAt,
			},
			IsSample: testcase.IsSample,
			Command:  testcase.Command,
		})
	}

	depends := make([]*dto.DependencyTestcase, 0, len(gradingItem.Depends))
	for _, depend := range gradingItem.Depends {
		depends = append(depends, &dto.DependencyTestcase{
			Testcase: dto.DependencyTestcaseInfo{
				ID:   depend.ID,
				Name: depend.Name,
			},
			GradingItem: dto.GradingItemInfo{
				ID:    depend.GradingItemID,
				Name:  depend.GradingItem.Name,
				Label: depend.GradingItem.Label,
			},
		})
	}

	resultParser := gradingItem.ResultParser
	if resultParser == "" {
		var project model.Project
		if err := q.db.Select("result_parser").Where("id = ?", gradingItem.ProjectID).First(&project).Error; err == nil {
			resultParser = project.ResultParser
		}
	}
	if resultParser == "" {
		resultParser = resultparser.AutoParser.String()
	}

	return &dto.GradingItemDetail{
		GradingItem: dto.GradingItem{
			ID:          gradingItem.ID,
			Name:        gradingItem.Name,
			Label:       gradingItem.Label,
			Description: gradingItem.Description,
			Order:       gradingItem.Order,
			Score:       gradingItem.Score,
			CreatedAt:   gradingItem.CreatedAt,
			UpdatedAt:   gradingItem.UpdatedAt,
		},
		ScoringStrategy: gradingItem.ScoringStrategy,
		OutputFormat:    gradingItem.OutputFormat,
		ResultParser:    resultParser,
		Testcases:       testcases,
		Dependencies:    depends,
	}, nil
}
