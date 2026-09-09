package query

import (
	"testing"

	sharedlogger "judge-backend/internal/app/shared/infrastructure/logger"
	"judge-backend/internal/app/shared/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// sampleTestcaseIDsForContestProject feeds both the REST detail's testcase
// filter and the stream's — it decides whose run output a contestant sees, so
// its per-problem scoping is pinned directly. (It replaced a global is_sample
// pluck that admitted every problem's sample testcase ids.)
func TestSampleTestcaseIDsForContestProject(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Discard,
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&model.ContestProject{}, &model.Project{}, &model.GradingItem{}, &model.Testcase{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	// Two problems, each with a sample and a hidden testcase; problem 1 also
	// has a sample under a soft-deleted grading item.
	must(db.Create(&model.Project{ID: 1, Name: "P1"}).Error)
	must(db.Create(&model.Project{ID: 2, Name: "P2"}).Error)
	must(db.Create(&model.ContestProject{ID: 10, ContestID: 1, ProjectID: 1, Name: "CP10"}).Error)
	must(db.Create(&model.ContestProject{ID: 20, ContestID: 1, ProjectID: 2, Name: "CP20"}).Error)
	must(db.Create(&model.GradingItem{ID: 1, ProjectID: 1, Name: "own", ScoringStrategy: "increase", OutputFormat: "text"}).Error)
	must(db.Create(&model.GradingItem{ID: 2, ProjectID: 1, Name: "gone", ScoringStrategy: "increase", OutputFormat: "text"}).Error)
	must(db.Create(&model.GradingItem{ID: 9, ProjectID: 2, Name: "other", ScoringStrategy: "increase", OutputFormat: "text"}).Error)
	must(db.Create(&model.Testcase{ID: 1, GradingItemID: 1, Name: "own-sample", Command: "true", IsSample: true}).Error)
	must(db.Create(&model.Testcase{ID: 2, GradingItemID: 1, Name: "own-hidden", Command: "true", IsSample: false}).Error)
	must(db.Create(&model.Testcase{ID: 3, GradingItemID: 2, Name: "orphaned-sample", Command: "true", IsSample: true}).Error)
	must(db.Create(&model.Testcase{ID: 9, GradingItemID: 9, Name: "other-sample", Command: "true", IsSample: true}).Error)
	must(db.Delete(&model.GradingItem{}, "id = ?", 2).Error)

	q := &SubmissionQuery{db: db, logger: sharedlogger.NewMockLogger()}

	ids, err := q.sampleTestcaseIDsForContestProject(10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 1 || ids[0] != 1 {
		t.Fatalf("expected exactly [1] (own live sample; not the other problem's, not the deleted item's), got %v", ids)
	}

	ids, err = q.sampleTestcaseIDsForContestProject(20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 1 || ids[0] != 9 {
		t.Fatalf("expected exactly [9] for the second problem, got %v", ids)
	}
}
