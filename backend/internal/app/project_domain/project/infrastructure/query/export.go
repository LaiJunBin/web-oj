package query

import (
	"errors"

	"judge-backend/internal/app/project_domain/project/application/query"
	"judge-backend/internal/app/project_domain/project/application/query/dto"
	"judge-backend/internal/app/shared/model"

	"gorm.io/gorm"
)

type ProjectExportQuery struct {
	db *gorm.DB
}

func NewProjectExportQuery(db *gorm.DB) query.ProjectExportQuery {
	return &ProjectExportQuery{db: db}
}

// LoadProjectForExport reads the whole project in one pass.
//
// Every collection is ordered explicitly, including the dependency joins.
// Without an ORDER BY, the database is free to return rows however it likes,
// and two exports of an unchanged project would differ — which defeats the
// point of a bundle you can diff in review. The unit tests cannot catch this
// because they build the struct directly; only an explicit order in the query
// makes the guarantee real.
func (q ProjectExportQuery) LoadProjectForExport(id int32) (*dto.ProjectExport, error) {
	m := model.Project{}
	err := q.db.
		Preload("DynamicFields", orderByID).
		Preload("Lifecycles", orderByID).
		Preload("Assets", func(db *gorm.DB) *gorm.DB { return db.Order("id ASC") }).
		Preload("GradingItems", func(db *gorm.DB) *gorm.DB {
			return db.Order("`order` ASC, id ASC")
		}).
		Preload("GradingItems.Testcases", orderByID).
		Preload("GradingItems.Testcases.Depends", orderByID).
		Preload("GradingItems.Depends", orderByID).
		Where("id = ?", id).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, query.ErrorProjectNotFound
	}
	if err != nil {
		return nil, err
	}

	out := &dto.ProjectExport{
		ID:           m.ID,
		Name:         m.Name,
		Description:  m.Description,
		Env:          m.Env,
		TestEnv:      m.TestEnv,
		ResultParser: m.ResultParser,
	}
	if m.ImageTag != nil {
		out.ImageTag = *m.ImageTag
	}
	if m.TestImageTag != nil {
		out.TestImageTag = *m.TestImageTag
	}

	for _, f := range m.DynamicFields {
		out.DynamicFields = append(out.DynamicFields, dto.ExportDynamicField{
			Type:         f.Type,
			Name:         f.Name,
			Label:        f.Label,
			DefaultValue: f.DefaultValue,
			IsRequired:   f.IsRequired,
		})
	}

	for _, l := range m.Lifecycles {
		out.Lifecycles = append(out.Lifecycles, dto.ExportLifecycle{
			Name:        l.Name,
			Description: l.Description,
			Type:        l.Type,
			RunsOn:      l.RunsOn,
			Command:     l.Command,
		})
	}

	for _, g := range m.GradingItems {
		item := dto.ExportGradingItem{
			ID:                g.ID,
			Label:             g.Label,
			Name:              g.Name,
			Description:       g.Description,
			Score:             g.Score,
			ScoringStrategy:   g.ScoringStrategy,
			OutputFormat:      g.OutputFormat,
			ResultParser:      g.ResultParser,
			DependTestcaseIDs: testcaseIDs(g.Depends),
		}
		for _, t := range g.Testcases {
			item.Testcases = append(item.Testcases, dto.ExportTestcase{
				ID:                t.ID,
				Name:              t.Name,
				Command:           t.Command,
				Score:             t.Score,
				IsSample:          t.IsSample,
				DependTestcaseIDs: testcaseIDs(t.Depends),
			})
		}
		out.GradingItems = append(out.GradingItems, item)
	}

	for _, a := range m.Assets {
		out.Assets = append(out.Assets, dto.ExportAsset{
			ID:          a.ID,
			Name:        a.Name,
			Description: a.Description,
			IsEnabled:   a.IsEnabled,
		})
	}

	return out, nil
}

func orderByID(db *gorm.DB) *gorm.DB {
	return db.Order("id ASC")
}

func testcaseIDs(ts []*model.Testcase) []int32 {
	ids := make([]int32, 0, len(ts))
	for _, t := range ts {
		ids = append(ids, t.ID)
	}
	return ids
}
