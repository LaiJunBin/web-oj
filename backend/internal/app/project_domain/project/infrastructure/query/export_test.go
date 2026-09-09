package query

import (
	"testing"

	"judge-backend/internal/app/shared/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// LoadProjectForExport is nine preloads deep and crosses both dependency join
// tables, one of which is self-referential. A feature test covers it only
// through an HTTP round trip, which makes a failure here look like an export
// bug rather than a query bug — this file is what told the difference the one
// time it mattered.
func exportDB(t *testing.T) *gorm.DB {
	t.Helper()

	// Each test gets its own in-memory database, so one leaving rows behind
	// cannot decide what another sees.
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Discard,
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.AutoMigrate(
		&model.Image{}, &model.Project{}, &model.DynamicField{}, &model.Lifecycle{},
		&model.GradingItem{}, &model.Testcase{}, &model.Asset{},
		&model.GradingItemDepend{}, &model.TestcaseDepend{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func seedExportFixture(t *testing.T, db *gorm.DB) {
	t.Helper()
	tag := "nginx:latest"

	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("seed: %v", err)
		}
	}

	must(db.Create(&model.Project{ID: 11, Name: "P", ImageTag: &tag, TestImageTag: &tag}).Error)
	// Display order and id order DISAGREE on purpose: the lower id is the
	// second item. Without that the assertion below proves nothing — a table
	// scan returns rows by id, which would look identical to a correct ORDER
	// BY, and dropping the ordering would go unnoticed. (It did: the first
	// version of this fixture had them agree and the check passed with the
	// ORDER BY deleted.)
	must(db.Create(&model.GradingItem{ID: 41, ProjectID: 11, Name: "second", Label: "2", Order: 2, ScoringStrategy: "increase", OutputFormat: "html"}).Error)
	must(db.Create(&model.GradingItem{ID: 42, ProjectID: 11, Name: "first", Label: "1", Order: 1, ScoringStrategy: "decrease", OutputFormat: "text"}).Error)
	must(db.Create(&model.Testcase{ID: 51, GradingItemID: 42, Name: "t1", Command: "true"}).Error)
	must(db.Create(&model.Testcase{ID: 52, GradingItemID: 42, Name: "t2", Command: "true"}).Error)
	must(db.Create(&model.Testcase{ID: 53, GradingItemID: 41, Name: "t3", Command: "true"}).Error)
	must(db.Create(&model.Lifecycle{ID: 31, ProjectID: 11, Name: "lc", Type: "before_all", RunsOn: "project", Command: "true"}).Error)
	must(db.Create(&model.DynamicField{ID: 21, ProjectID: 11, Type: "text", Name: "f", Label: "F"}).Error)
	must(db.Create(&model.Asset{ID: "a-1", ProjectID: 11, Name: "spec.pdf", Description: "d", Size: 3, IsEnabled: true}).Error)

	// The two dependency shapes that a naive query gets wrong: a testcase
	// depending on one under ANOTHER grading item, and a grading item
	// depending on a testcase.
	must(db.Create(&model.TestcaseDepend{TestcaseID: 53, DependID: 51}).Error)
	must(db.Create(&model.GradingItemDepend{GradingItemID: 41, TestcaseID: 52}).Error)
}

func TestLoadProjectForExportReadsTheWholeAggregate(t *testing.T) {
	db := exportDB(t)
	seedExportFixture(t, db)

	got, err := NewProjectExportQuery(db).LoadProjectForExport(11)
	if err != nil {
		t.Fatalf("LoadProjectForExport: %v", err)
	}

	if len(got.GradingItems) != 2 {
		t.Fatalf("expected 2 grading items, got %d", len(got.GradingItems))
	}
	if len(got.Lifecycles) != 1 || len(got.DynamicFields) != 1 || len(got.Assets) != 1 {
		t.Errorf("expected one of each child, got %d lifecycles, %d fields, %d assets",
			len(got.Lifecycles), len(got.DynamicFields), len(got.Assets))
	}

	// Every collection is ordered explicitly in the query. Without that the
	// database may return rows however it likes and two exports of an
	// unchanged project would differ — which defeats a bundle meant to be
	// diffed. The fixture gives the lower id the higher display order, so a
	// dropped ORDER BY shows up here instead of looking correct by accident.
	if got.GradingItems[0].Name != "first" || got.GradingItems[1].Name != "second" {
		t.Errorf("expected grading items in display order, got %q then %q",
			got.GradingItems[0].Name, got.GradingItems[1].Name)
	}

	// Description exists on the asset row but not on the entity, which is why
	// the export needs its own query rather than the screen's DTO.
	if got.Assets[0].Description != "d" {
		t.Errorf("expected the asset description to survive, got %q", got.Assets[0].Description)
	}
}

func TestLoadProjectForExportResolvesBothDependencyShapes(t *testing.T) {
	db := exportDB(t)
	seedExportFixture(t, db)

	got, err := NewProjectExportQuery(db).LoadProjectForExport(11)
	if err != nil {
		t.Fatalf("LoadProjectForExport: %v", err)
	}

	second := got.GradingItems[1]
	if len(second.DependTestcaseIDs) != 1 || second.DependTestcaseIDs[0] != 52 {
		t.Errorf("expected grading item 41 to depend on testcase 52, got %v", second.DependTestcaseIDs)
	}

	// The cross-item case: t3 lives under grading item 42 and depends on t1,
	// which lives under 41. A key scheme scoped per grading item would lose
	// this, so it is worth pinning at the query as well as in the bundle.
	tc := second.Testcases[0]
	if len(tc.DependTestcaseIDs) != 1 || tc.DependTestcaseIDs[0] != 51 {
		t.Errorf("expected testcase 53 to depend on testcase 51, got %v", tc.DependTestcaseIDs)
	}
}

func TestLoadProjectForExportReportsMissingProject(t *testing.T) {
	db := exportDB(t)

	_, err := NewProjectExportQuery(db).LoadProjectForExport(999)

	// Not (nil, nil): the older queries in this package use that spelling and
	// it puts the burden on every caller to remember a nil result is not a bug.
	if err == nil {
		t.Fatal("expected an error for a project that does not exist")
	}
}
