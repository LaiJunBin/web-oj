package query

import (
	"errors"
	"testing"

	sharedlogger "judge-backend/internal/app/shared/infrastructure/logger"
	"judge-backend/internal/app/shared/model"
	"judge-backend/internal/app/shared/types/httperror"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// CheckProjectAndGradingItemsExists is the ownership gate for partial
// submissions; a feature test only sees it through an HTTP round trip, so a
// hole here (foreign or ghost ids sliding through) looks like a judge bug
// rather than a query bug. This file pins the set semantics directly.
func contestProjectDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Discard,
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.AutoMigrate(
		&model.Contest{}, &model.ContestProject{}, &model.Project{}, &model.GradingItem{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

// seedTwoProblems wires two problems under one contest: contest project 10
// serves project 1 (live grading items 1 and 2, soft-deleted item 3), contest
// project 20 serves project 2 (grading item 9). Item 9 is the "foreign" id
// from contest project 10's point of view.
func seedTwoProblems(t *testing.T, db *gorm.DB) {
	t.Helper()
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("seed: %v", err)
		}
	}

	must(db.Create(&model.Project{ID: 1, Name: "P1"}).Error)
	must(db.Create(&model.Project{ID: 2, Name: "P2"}).Error)
	must(db.Create(&model.ContestProject{ID: 10, ContestID: 1, ProjectID: 1, Name: "CP10"}).Error)
	must(db.Create(&model.ContestProject{ID: 20, ContestID: 1, ProjectID: 2, Name: "CP20"}).Error)
	must(db.Create(&model.GradingItem{ID: 1, ProjectID: 1, Name: "own-1", ScoringStrategy: "increase", OutputFormat: "text"}).Error)
	must(db.Create(&model.GradingItem{ID: 2, ProjectID: 1, Name: "own-2", ScoringStrategy: "increase", OutputFormat: "text"}).Error)
	must(db.Create(&model.GradingItem{ID: 3, ProjectID: 1, Name: "deleted", ScoringStrategy: "increase", OutputFormat: "text"}).Error)
	must(db.Delete(&model.GradingItem{}, "id = ?", 3).Error)
	must(db.Create(&model.GradingItem{ID: 9, ProjectID: 2, Name: "foreign", ScoringStrategy: "increase", OutputFormat: "text"}).Error)
}

func newContestProjectQueryForTest(t *testing.T) (*ContestProjectQuery, *gorm.DB) {
	t.Helper()
	db := contestProjectDB(t)
	seedTwoProblems(t, db)
	return &ContestProjectQuery{db: db, logger: sharedlogger.NewMockLogger()}, db
}

func assertBadGradingItems(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	he := &httperror.Error{}
	if !errors.As(err, &he) {
		t.Fatalf("expected *httperror.Error, got %T: %v", err, err)
	}
	if he.Code != 400 || he.Message != "invalid grading item ids" {
		t.Fatalf("expected 400 %q, got %d %q", "invalid grading item ids", he.Code, he.Message)
	}
}

func TestCheckProjectAndGradingItemsExists(t *testing.T) {
	t.Run("an empty list is a whole-project submission and always passes", func(t *testing.T) {
		q, db := newContestProjectQueryForTest(t)
		// Even for a problem with zero grading items: the empty list means
		// "everything", it is not a claim about specific ids.
		if err := db.Create(&model.ContestProject{ID: 30, ContestID: 1, ProjectID: 1, Name: "CP30"}).Error; err != nil {
			t.Fatal(err)
		}
		if err := q.CheckProjectAndGradingItemsExists(10, []int32{}); err != nil {
			t.Fatalf("empty list: %v", err)
		}
		if err := q.CheckProjectAndGradingItemsExists(10, nil); err != nil {
			t.Fatalf("nil list: %v", err)
		}
	})

	t.Run("ids owned by the problem pass", func(t *testing.T) {
		q, _ := newContestProjectQueryForTest(t)
		if err := q.CheckProjectAndGradingItemsExists(10, []int32{1, 2}); err != nil {
			t.Fatalf("owned ids: %v", err)
		}
	})

	t.Run("duplicate ids collapse instead of failing the set comparison", func(t *testing.T) {
		q, _ := newContestProjectQueryForTest(t)
		if err := q.CheckProjectAndGradingItemsExists(10, []int32{1, 1, 2}); err != nil {
			t.Fatalf("duplicated ids: %v", err)
		}
	})

	t.Run("an id belonging to another problem is rejected", func(t *testing.T) {
		q, _ := newContestProjectQueryForTest(t)
		assertBadGradingItems(t, q.CheckProjectAndGradingItemsExists(10, []int32{1, 9}))
	})

	t.Run("a set made up entirely of another problem's ids is rejected", func(t *testing.T) {
		q, _ := newContestProjectQueryForTest(t)
		assertBadGradingItems(t, q.CheckProjectAndGradingItemsExists(10, []int32{9}))
	})

	t.Run("a nonexistent id is rejected", func(t *testing.T) {
		q, _ := newContestProjectQueryForTest(t)
		assertBadGradingItems(t, q.CheckProjectAndGradingItemsExists(10, []int32{1, 999}))
	})

	t.Run("a soft-deleted id is rejected for new submissions", func(t *testing.T) {
		// The judge side stays lenient about deleted items on OLD submissions;
		// a NEW submission has no reason to reference one.
		q, _ := newContestProjectQueryForTest(t)
		assertBadGradingItems(t, q.CheckProjectAndGradingItemsExists(10, []int32{3}))
	})

	t.Run("a missing contest project keeps its 404", func(t *testing.T) {
		q, _ := newContestProjectQueryForTest(t)
		err := q.CheckProjectAndGradingItemsExists(999, []int32{})
		he := &httperror.Error{}
		if !errors.As(err, &he) || he.Code != 404 || he.Message != "contest project not found" {
			t.Fatalf("expected 404 contest project not found, got %v", err)
		}
	})
}
