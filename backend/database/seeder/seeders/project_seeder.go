package seeders

import (
	"judge-backend/internal/app/shared/model"

	"gorm.io/gorm"
)

type ProjectSeeder struct{}

func stringPtr(s string) *string {
	return &s
}

func (s ProjectSeeder) Seed(db *gorm.DB) error {
	data, err := LoadDemoData()
	if err != nil {
		return err
	}

	// The demo project runs the contestant's files in an nginx container and
	// grades them from a second container of the same image, so both tags
	// point at the one image the dataset seeds.
	project := &model.Project{
		Name:         data.Project.Name,
		Description:  data.Project.Description,
		ImageTag:     stringPtr(data.Image.Tag),
		TestImageTag: stringPtr(data.Image.Tag),
		ResultParser: data.Project.ResultParser,
	}

	if err := db.Create(project).Error; err != nil {
		return err
	}

	for _, field := range data.Project.DynamicFields {
		if err := db.Create(&model.DynamicField{
			ProjectID:    project.ID,
			Type:         field.Type,
			Name:         field.Name,
			Label:        field.Label,
			DefaultValue: field.DefaultValue,
			IsRequired:   field.IsRequired,
		}).Error; err != nil {
			return err
		}
	}

	for _, lc := range data.Project.Lifecycles {
		if err := db.Create(&model.Lifecycle{
			ProjectID:   project.ID,
			Name:        lc.Name,
			Description: lc.Description,
			Type:        lc.Type,
			RunsOn:      lc.RunsOn,
			Command:     lc.Command,
		}).Error; err != nil {
			return err
		}
	}

	// Order follows the dataset's order rather than the label, so reordering
	// the JSON is enough to reorder the grading sheet.
	for i, item := range data.Project.GradingItems {
		gradingItem := &model.GradingItem{
			ProjectID:       project.ID,
			Name:            item.Name,
			Description:     item.Description,
			Score:           item.Score,
			Order:           uint32(i + 1),
			ScoringStrategy: item.ScoringStrategy,
			OutputFormat:    item.OutputFormat,
			Label:           item.Label,
		}

		if err := db.Create(gradingItem).Error; err != nil {
			return err
		}

		for _, tc := range item.Testcases {
			if err := db.Create(&model.Testcase{
				GradingItemID: gradingItem.ID,
				Name:          tc.Name,
				Command:       tc.Command,
				Score:         tc.Score,
				IsSample:      tc.IsSample,
			}).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
