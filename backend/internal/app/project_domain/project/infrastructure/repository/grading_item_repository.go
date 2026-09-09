package repository

import (
	"judge-backend/internal/app/shared/model"
	"judge-backend/internal/app/project_domain/project/domain/entity/gradingitem"
	"judge-backend/internal/app/project_domain/project/domain/entity/project"
	"judge-backend/internal/app/project_domain/project/domain/repository"

	"gorm.io/gorm"
)

type GradingItemRepository struct {
	db *gorm.DB
}

func NewGradingItemRepository(db *gorm.DB) repository.GradingItemRepository {
	return &GradingItemRepository{
		db: db,
	}
}

func (r *GradingItemRepository) toDomainGradingItem(g *model.GradingItem) (*gradingitem.GradingItem, error) {
	d, err := gradingitem.New(&gradingitem.GradingItemConfig{
		ProjectID:       project.ID(g.ProjectID),
		Name:            g.Name,
		Description:     g.Description,
		Order:           g.Order,
		Score:           g.Score,
		ScoringStrategy: g.ScoringStrategy,
		OutputFormat:    g.OutputFormat,
		ResultParser:    g.ResultParser,
		Label:           g.Label,
		DependTestcases: func() []int32 {
			depends := make([]int32, 0, len(g.Depends))
			for _, dep := range g.Depends {
				depends = append(depends, dep.ID)
			}

			return depends
		}(),
	})
	if err != nil {
		return nil, err
	}

	d.SetID(g.ID)
	d.SetCreatedAt(g.CreatedAt)
	d.SetUpdatedAt(g.UpdatedAt)

	return d, nil
}

func (r *GradingItemRepository) Exists(id int) (bool, error) {
	var count int64
	if err := r.db.Model(&model.GradingItem{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *GradingItemRepository) FindByID(id int) (*gradingitem.GradingItem, error) {
	gradingItem := &model.GradingItem{}
	if err := r.db.Preload("Depends").First(gradingItem, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, err
	}

	return r.toDomainGradingItem(gradingItem)
}

func (r *GradingItemRepository) Save(gradingItem *gradingitem.GradingItem) error {
	depends := make([]*model.Testcase, 0, len(gradingItem.DependTestcases()))
	for _, id := range gradingItem.DependTestcases() {
		dep := &model.Testcase{}
		if err := r.db.First(dep, "id = ?", id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return repository.ErrorDependTestcaseNotFound
			}

			return err
		}

		if dep.GradingItemID == gradingItem.ID() {
			return repository.ErrorDependTestcaseIsSameGradingItem
		}

		depends = append(depends, &model.Testcase{
			ID: id,
		})
	}

	g := &model.GradingItem{
		ID:              gradingItem.ID(),
		ProjectID:       gradingItem.ProjectID().Value(),
		Name:            gradingItem.Name().String(),
		Score:           gradingItem.Score(),
		Order:           gradingItem.Order(),
		Description:     gradingItem.Description(),
		ScoringStrategy: gradingItem.ScoringStrategy().String(),
		OutputFormat:    gradingItem.OutputFormater().String(),
		ResultParser:    gradingItem.ResultParser(),
		Label:           gradingItem.Label(),
		Depends:         depends,
	}

	if gradingItem.ID() == 0 {
		var maxOrder *uint32
		if err := r.db.Model(&model.GradingItem{}).Where("project_id = ?", gradingItem.ProjectID().Value()).Select("max(`order`)").Row().Scan(&maxOrder); err != nil {
			return err
		}

		if maxOrder == nil {
			g.Order = 1
		} else {
			g.Order = *maxOrder + 1
		}
	}

	if !gradingItem.CreatedAt().IsZero() {
		g.CreatedAt = gradingItem.CreatedAt()
	}

	if !gradingItem.UpdatedAt().IsZero() {
		g.UpdatedAt = gradingItem.UpdatedAt()
	}

	var oldOrder *uint32
	if gradingItem.ID() != 0 {
		if err := r.db.First(&model.GradingItem{}, "id = ?", gradingItem.ID()).Select("`order`").Row().Scan(&oldOrder); err != nil {
			return err
		}
	}

	tx := r.db.Begin()
	if g.ID != 0 {
		if err := tx.Model(g).Association("Depends").Replace(depends); err != nil {
			tx.Rollback()
			return err
		}
	}

	if oldOrder != nil && *oldOrder != g.Order {
		if *oldOrder < g.Order {
			if err := tx.Exec("UPDATE grading_items SET `order` = `order` - 1 WHERE project_id = ? AND `order` > ? AND `order` <= ?", g.ProjectID, *oldOrder, g.Order).Error; err != nil {
				tx.Rollback()
				return err
			}
		} else {
			if err := tx.Exec("UPDATE grading_items SET `order` = `order` + 1 WHERE project_id = ? AND `order` < ? AND `order` >= ?", g.ProjectID, *oldOrder, g.Order).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	if err := tx.Save(g).Error; err != nil {
		tx.Rollback()
		return err
	}

	gradingItem.SetID(g.ID)

	return tx.Commit().Error
}

func (r *GradingItemRepository) Delete(gradingItem *gradingitem.GradingItem) error {
	return r.db.Delete(&model.GradingItem{}, "id = ?", gradingItem.ID()).Error
}
