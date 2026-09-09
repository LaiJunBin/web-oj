package repository

import (
	"judge-backend/internal/app/shared/model"
	"judge-backend/internal/app/project_domain/project/domain/entity/testcase"
	"judge-backend/internal/app/project_domain/project/domain/repository"

	"gorm.io/gorm"
)

type TestcaseRepository struct {
	db *gorm.DB
}

func NewTestcaseRepository(db *gorm.DB) repository.TestcaseRepository {
	return &TestcaseRepository{
		db: db,
	}
}

func (r *TestcaseRepository) Exists(id int) (bool, error) {
	var count int64
	if err := r.db.Model(&model.Testcase{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *TestcaseRepository) toDomainModel(tc *model.Testcase) (*testcase.Testcase, error) {
	t, err := testcase.New(&testcase.TestcaseConfig{
		GradingItemID: tc.GradingItemID,
		Name:          tc.Name,
		Command:       tc.Command,
		Score:         tc.Score,
		IsSample:      tc.IsSample,
		Depends: func() []int32 {
			depends := make([]int32, 0, len(tc.Depends))
			for _, dep := range tc.Depends {
				depends = append(depends, dep.ID)
			}
			return depends
		}(),
	})
	if err != nil {
		return nil, err
	}

	t.SetID(tc.ID)
	t.SetCreatedAt(tc.CreatedAt)
	t.SetUpdatedAt(tc.UpdatedAt)

	return t, nil
}

func (r *TestcaseRepository) FindByID(id int) (*testcase.Testcase, error) {
	testcase := &model.Testcase{}
	if err := r.db.Preload("Depends").First(&testcase, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return r.toDomainModel(testcase)
}

func (r *TestcaseRepository) Save(testcase *testcase.Testcase) error {
	depends := make([]*model.Testcase, 0, len(testcase.Depends()))
	for _, id := range testcase.Depends() {
		dep := &model.Testcase{}
		if err := r.db.First(dep, "id = ?", id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return repository.ErrorDependTestcaseNotFound
			}

			return err
		}

		depends = append(depends, &model.Testcase{
			ID: id,
		})
	}

	t := &model.Testcase{
		ID:            testcase.ID(),
		GradingItemID: testcase.GradingItemID(),
		Name:          testcase.Name().String(),
		Score:         testcase.Score(),
		Command:       testcase.Command(),
		IsSample:      testcase.IsSample(),
		Depends:       depends,
	}

	if !testcase.CreatedAt().IsZero() {
		t.CreatedAt = testcase.CreatedAt()
	}

	tx := r.db.Begin()
	if t.ID != 0 {
		if err := tx.Model(t).Association("Depends").Replace(depends); err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := tx.Save(t).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}

	testcase.SetID(t.ID)
	testcase.SetCreatedAt(t.CreatedAt)
	testcase.SetUpdatedAt(t.UpdatedAt)
	return nil
}

func (r *TestcaseRepository) Delete(testcase *testcase.Testcase) error {
	t := &model.Testcase{
		ID: testcase.ID(),
	}
	if err := r.db.Delete(t).Error; err != nil {
		return err
	}
	return nil
}
