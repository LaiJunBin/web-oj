package seeder_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"judge-backend/database/migration"
	"judge-backend/database/seeder"
	"judge-backend/database/seeder/seeders"
	"judge-backend/internal/app/shared/model"
	"judge-backend/internal/config"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// The demo dataset is what a fresh clone shows on first login and what the
// operating manual screenshots, so these tests guard the two promises that
// make it useful: every account can log in and see a contest that is running
// right now, and the exercise is gradeable end to end.

func seedDB(t *testing.T) *gorm.DB {
	t.Helper()

	// A shared cache keyed on the test name: plain "file::memory:" gives every
	// pooled connection its own empty database, so the migrations would
	// intermittently appear to be missing. The per-test key keeps the three
	// tests from seeding into each other.
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	migration.RunMigrate(db)
	seeder.RunSeed(db, config.NewTestConfig())
	return db
}

func TestSeedCreatesLoginableDemoAccounts(t *testing.T) {
	db := seedDB(t)
	data, err := seeders.LoadDemoData()
	if err != nil {
		t.Fatalf("load demo data: %v", err)
	}

	for _, want := range data.Users() {
		var user model.User
		if err := db.Where("username = ?", want.Username).First(&user).Error; err != nil {
			t.Fatalf("user %q not seeded: %v", want.Username, err)
		}
		if user.Name != want.Name {
			t.Errorf("user %q name = %q, want %q", want.Username, user.Name, want.Name)
		}
		// Stored hashed, never in the clear — the shared demo password is
		// public, the column still must not be.
		if user.Password == data.Password {
			t.Errorf("user %q password stored in plaintext", want.Username)
		}
	}
}

func TestSeededContestIsRunningAndFullyEnrolled(t *testing.T) {
	db := seedDB(t)
	data, err := seeders.LoadDemoData()
	if err != nil {
		t.Fatalf("load demo data: %v", err)
	}

	var contest model.Contest
	if err := db.Preload("Projects").Where("name = ?", data.Contest.Name).First(&contest).Error; err != nil {
		t.Fatalf("demo contest not seeded: %v", err)
	}

	if !contest.IsActivated() {
		t.Errorf("demo contest is not active at seed time (activate %v, deactivate %v)",
			contest.ActivateTime, contest.DeactivateTime)
	}

	if len(contest.Projects) != 1 {
		t.Fatalf("contest projects = %d, want 1", len(contest.Projects))
	}

	cp := contest.Projects[0]
	if status := cp.Status(); status != "InProgress" {
		t.Errorf("contest project status = %q, want InProgress", status)
	}

	// The papercut this dataset replaces: a contest that ended minutes after
	// installation.
	if left := time.Until(cp.EndTime); left < 24*time.Hour {
		t.Errorf("contest project ends in %v, too soon to demo with", left)
	}

	var enrolled int64
	if err := db.Model(&model.ContestParticipant{}).Where("contest_id = ?", contest.ID).Count(&enrolled).Error; err != nil {
		t.Fatalf("count participants: %v", err)
	}
	if want := int64(len(data.Usernames())); enrolled != want {
		t.Errorf("enrolled participants = %d, want %d", enrolled, want)
	}
}

func TestSeededProjectIsGradeable(t *testing.T) {
	db := seedDB(t)
	data, err := seeders.LoadDemoData()
	if err != nil {
		t.Fatalf("load demo data: %v", err)
	}

	var project model.Project
	if err := db.Preload("GradingItems.Testcases").
		Where("name = ?", data.Project.Name).First(&project).Error; err != nil {
		t.Fatalf("demo project not seeded: %v", err)
	}

	if project.ImageTag == nil || *project.ImageTag != data.Image.Tag {
		t.Fatalf("project image tag = %v, want %q", project.ImageTag, data.Image.Tag)
	}

	// The image the project runs on has to be seeded too, or the first
	// judgment fails on a dangling tag.
	var image model.Image
	if err := db.Where("tag = ?", data.Image.Tag).First(&image).Error; err != nil {
		t.Fatalf("image %q not seeded: %v", data.Image.Tag, err)
	}

	if len(project.GradingItems) != len(data.Project.GradingItems) {
		t.Fatalf("grading items = %d, want %d", len(project.GradingItems), len(data.Project.GradingItems))
	}

	var total float64
	for _, item := range project.GradingItems {
		if len(item.Testcases) == 0 {
			t.Errorf("grading item %q has no testcases", item.Name)
		}
		total += item.Score
	}

	// 25 points is the figure the manual quotes alongside the 18/25 the
	// bundled sample submission earns.
	if total != 25 {
		t.Errorf("total score = %g, want 25", total)
	}
}

// TestNegatedGrepsGuardFileExistence pins down a footgun that silently inverts
// a grading item. `! grep -q PATTERN missing-file` succeeds: grep exits 2
// because it cannot open the file, and `!` turns that into a pass. Any
// testcase that excludes content therefore has to establish the file exists
// first, or not submitting the file at all scores full marks.
func TestNegatedGrepsGuardFileExistence(t *testing.T) {
	data, err := seeders.LoadDemoData()
	if err != nil {
		t.Fatalf("load demo data: %v", err)
	}

	for _, item := range data.Project.GradingItems {
		for _, tc := range item.Testcases {
			neg := strings.Index(tc.Command, "! grep")
			if neg < 0 {
				continue
			}
			if guard := strings.Index(tc.Command, "test -f"); guard < 0 || guard > neg {
				t.Errorf("grading item %q testcase %q negates grep without a preceding `test -f`; "+
					"a missing file would pass:\n  %s", item.Name, tc.Name, tc.Command)
			}
		}
	}
}
