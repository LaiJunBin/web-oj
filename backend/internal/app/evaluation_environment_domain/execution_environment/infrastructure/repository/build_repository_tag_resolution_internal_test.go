package repository

import (
	"testing"
	"time"

	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build"
	"judge-backend/internal/app/shared/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// A force import replaces an image by soft-deleting the old row and inserting
// a new one, so one tag then maps to two rows. Every function here that
// resolves a tag has to pick the live row — the first version resolved with an
// unscoped unordered First, took the dead row, and attached every subsequent
// build to it: invisible in the builds list, and a 500 from rebuild and
// update-with-context (whose FindLastBuild walked the dead image's builds into
// a scoped lookup that could not see it).
func resolutionDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Discard,
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.Image{}, &model.Build{}, &model.BuildLog{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func forceImportedTag(t *testing.T, db *gorm.DB) (old, live model.Image) {
	t.Helper()

	// The dead row has the LOWER id, exactly as after a real force import —
	// that is the row an unordered unscoped First returns.
	old = model.Image{ID: 1, Tag: "forced:v1", MountPath: "/var/www/html"}
	old.DeletedAt = gorm.DeletedAt{Time: time.Now(), Valid: true}
	if err := db.Create(&old).Error; err != nil {
		t.Fatal(err)
	}

	live = model.Image{ID: 2, Tag: "forced:v1", MountPath: "/var/www/html"}
	if err := db.Create(&live).Error; err != nil {
		t.Fatal(err)
	}

	// The old incarnation had built twice. GetMaxBatch resolving the dead row
	// would hand the next build batch 3 instead of 1 — harmless-looking, which
	// is why the resolution itself is what gets pinned.
	for batch := 1; batch <= 2; batch++ {
		if err := db.Create(&model.Build{ImageID: old.ID, Batch: batch, Status: 2}).Error; err != nil {
			t.Fatal(err)
		}
	}
	return old, live
}

func TestTagResolutionPrefersTheLiveImage(t *testing.T) {
	db := resolutionDB(t)
	_, live := forceImportedTag(t, db)
	repo := &BuildRepository{db: db, buildMap: map[string]*build.Build{}}

	id, err := repo.tagToImageID("forced:v1")
	if err != nil {
		t.Fatalf("tagToImageID: %v", err)
	}
	if id != live.ID {
		t.Errorf("expected the live image %d, but resolved %d — builds would attach to the soft-deleted image", live.ID, id)
	}

	// Batch numbering restarts with the new incarnation: its builds are its
	// own, not a continuation of the dead image's.
	maxBatch, err := repo.GetMaxBatch("forced:v1")
	if err != nil {
		t.Fatalf("GetMaxBatch: %v", err)
	}
	if maxBatch != 0 {
		t.Errorf("expected max batch 0 for the fresh image, but got %d", maxBatch)
	}
}

// delete_image soft-deletes the row BEFORE cancelling its build, so during
// deletion the tag has no live row and must still resolve — to the newest
// dead incarnation, since that is the one being torn down.
func TestTagResolutionFallsBackToTheNewestDeletedImage(t *testing.T) {
	db := resolutionDB(t)
	old, live := forceImportedTag(t, db)
	// Soft-delete the live row too: mid-deletion state.
	if err := db.Delete(&model.Image{}, live.ID).Error; err != nil {
		t.Fatal(err)
	}
	repo := &BuildRepository{db: db, buildMap: map[string]*build.Build{}}

	id, err := repo.tagToImageID("forced:v1")
	if err != nil {
		t.Fatalf("tagToImageID: %v", err)
	}
	if id != live.ID {
		t.Errorf("expected the newest incarnation %d, but resolved %d", live.ID, id)
	}
	_ = old
}

// The other half of the original failure: a build row that belongs to a
// soft-deleted image must still convert. Scoped, this lookup turned the
// misattached rows into FindLastBuild errors, which is what escalated
// "invisible build" into "rebuild returns 500".
func TestImageIDToTagResolvesASoftDeletedImage(t *testing.T) {
	db := resolutionDB(t)
	old, _ := forceImportedTag(t, db)
	repo := &BuildRepository{db: db, buildMap: map[string]*build.Build{}}

	tag, err := repo.imageIDToTag(old.ID)
	if err != nil {
		t.Fatalf("imageIDToTag on a soft-deleted image: %v", err)
	}
	if tag != "forced:v1" {
		t.Errorf("expected forced:v1, got %q", tag)
	}
}
