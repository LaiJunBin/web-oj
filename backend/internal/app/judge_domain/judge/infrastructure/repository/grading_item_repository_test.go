package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	domainrepository "judge-backend/internal/app/judge_domain/judge/domain/repository"
	sharedlogger "judge-backend/internal/app/shared/infrastructure/logger"
	"judge-backend/internal/app/shared/model"

	"go.mongodb.org/mongo-driver/bson"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// These tests need BOTH stores: the grading-item set comes from MySQL/sqlite,
// but which ids a submission recorded lives in the Mongo submission document,
// and the ownership rules under test are exactly about the seam between the
// two. The Mongo container is the package-shared one from
// judgement_repository_mongo_test.go (skips without Docker).

func gradingItemDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Discard,
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&model.Project{}, &model.ContestProject{}, &model.GradingItem{}, &model.Testcase{},
		&model.GradingItemDepend{}, &model.TestcaseDepend{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

// seedGradingItemFixture wires two problems: contest project 10 serves
// project 1 (live grading items 1 and 2, soft-deleted item 3), contest
// project 20 serves project 2 (grading item 9 — the foreign id from contest
// project 10's point of view).
func seedGradingItemFixture(t *testing.T, db *gorm.DB) {
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
	must(db.Create(&model.Testcase{ID: 51, GradingItemID: 1, Name: "t1", Command: "true", Score: 10}).Error)
}

func newGradingItemRepoForTest(t *testing.T) *GradingItemRepository {
	t.Helper()
	db := gradingItemDB(t)
	seedGradingItemFixture(t, db)
	return &GradingItemRepository{
		db:     db,
		mdb:    mongoClientForTest(t),
		logger: sharedlogger.NewMockLogger(),
	}
}

// insertSubmissionDoc writes the Mongo submission document the repository
// resolves ids from. gradingItemIds is left as interface{} so the bad-type
// test can store something that is not an int32 array.
func insertSubmissionDoc(t *testing.T, r *GradingItemRepository, id string, contestProjectID int32, gradingItemIDs interface{}) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := r.mdb.Database("judge").Collection("submissions").InsertOne(ctx, bson.M{
		"_id":              id,
		"contestProjectID": contestProjectID,
		"gradingItemIds":   gradingItemIDs,
	})
	if err != nil {
		t.Fatalf("insert submission doc: %v", err)
	}
}

func TestGradingItemRepository_FindBySubmissionID(t *testing.T) {
	t.Run("an empty id list loads the whole problem", func(t *testing.T) {
		r := newGradingItemRepoForTest(t)
		insertSubmissionDoc(t, r, "sub-empty", 10, bson.A{})

		items, err := r.FindBySubmissionID("sub-empty")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(items) != 2 {
			t.Fatalf("expected the problem's 2 live grading items, got %d", len(items))
		}
	})

	t.Run("a legitimate partial set loads exactly the requested items", func(t *testing.T) {
		r := newGradingItemRepoForTest(t)
		insertSubmissionDoc(t, r, "sub-partial", 10, bson.A{int32(1)})

		items, err := r.FindBySubmissionID("sub-partial")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(items) != 1 || items[0].ID() != 1 {
			t.Fatalf("expected exactly grading item 1, got %v", items)
		}
	})

	t.Run("an id belonging to another problem fails the load", func(t *testing.T) {
		r := newGradingItemRepoForTest(t)
		insertSubmissionDoc(t, r, "sub-foreign", 10, bson.A{int32(1), int32(9)})

		_, err := r.FindBySubmissionID("sub-foreign")
		if !errors.Is(err, domainrepository.ErrGradingItemsNotOwned) {
			t.Fatalf("expected ErrGradingItemsNotOwned for a foreign grading item id, got %v", err)
		}
	})

	t.Run("an id that never existed fails the load", func(t *testing.T) {
		r := newGradingItemRepoForTest(t)
		insertSubmissionDoc(t, r, "sub-ghost", 10, bson.A{int32(1), int32(999)})

		_, err := r.FindBySubmissionID("sub-ghost")
		if !errors.Is(err, domainrepository.ErrGradingItemsNotOwned) {
			t.Fatalf("expected ErrGradingItemsNotOwned for a nonexistent grading item id, got %v", err)
		}
	})

	t.Run("a soft-deleted own item is skipped, not fatal", func(t *testing.T) {
		// Old submissions legitimately reference items an admin later deleted;
		// their rejudge keeps working exactly as before this validation.
		r := newGradingItemRepoForTest(t)
		insertSubmissionDoc(t, r, "sub-deleted", 10, bson.A{int32(1), int32(3)})

		items, err := r.FindBySubmissionID("sub-deleted")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(items) != 1 || items[0].ID() != 1 {
			t.Fatalf("expected only the live grading item 1, got %v", items)
		}
	})

	t.Run("a set of only soft-deleted own items resolves to empty", func(t *testing.T) {
		r := newGradingItemRepoForTest(t)
		insertSubmissionDoc(t, r, "sub-all-deleted", 10, bson.A{int32(3)})

		items, err := r.FindBySubmissionID("sub-all-deleted")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(items) != 0 {
			t.Fatalf("expected no items, got %v", items)
		}
	})

	t.Run("a malformed id array is an error, not a panic", func(t *testing.T) {
		r := newGradingItemRepoForTest(t)
		insertSubmissionDoc(t, r, "sub-bad-type", 10, bson.A{"not-an-int"})

		_, err := r.FindBySubmissionID("sub-bad-type")
		if err == nil {
			t.Fatal("expected an error for a malformed grading item id array, got none")
		}
	})
}
