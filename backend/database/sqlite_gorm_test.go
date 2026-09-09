package database

import (
	"judge-backend/internal/config"
	"testing"
)

// TestNewSqliteGormDB_SeedSmoke runs the exact migrate+seed path the sqlite
// dev server executes at startup, so a broken seeder fails here instead of
// crashing the dev server on its next restart. What the demo dataset should
// contain is asserted in database/seeder; this only checks that the boot path
// completes and lands the headline rows.
func TestNewSqliteGormDB_SeedSmoke(t *testing.T) {
	db := NewSqliteGormDB(config.NewTestConfig())

	var admins int64
	if err := db.Table("users").Where("username = ?", "admin").Count(&admins).Error; err != nil {
		t.Fatalf("query admin user: %v", err)
	}
	if admins != 1 {
		t.Errorf("expected exactly one admin account, got %d", admins)
	}

	var contests int64
	if err := db.Table("contests").Count(&contests).Error; err != nil {
		t.Fatalf("count contests: %v", err)
	}
	if contests != 1 {
		t.Errorf("expected the single demo contest, got %d", contests)
	}

	// Every testcase has to be attached to a grading item, or it silently
	// never runs.
	var orphanTestcases int64
	if err := db.Table("testcases").
		Where("grading_item_id NOT IN (SELECT id FROM grading_items)").
		Count(&orphanTestcases).Error; err != nil {
		t.Fatalf("count orphan testcases: %v", err)
	}
	if orphanTestcases != 0 {
		t.Errorf("expected no orphan testcases, got %d", orphanTestcases)
	}

	// Same for the project's image tag: a dangling tag fails the first
	// judgment rather than the seed.
	var danglingImages int64
	if err := db.Table("projects").
		Where("image_tag NOT IN (SELECT tag FROM images) OR test_image_tag NOT IN (SELECT tag FROM images)").
		Count(&danglingImages).Error; err != nil {
		t.Fatalf("count projects with dangling image tags: %v", err)
	}
	if danglingImages != 0 {
		t.Errorf("expected no projects with dangling image tags, got %d", danglingImages)
	}
}
