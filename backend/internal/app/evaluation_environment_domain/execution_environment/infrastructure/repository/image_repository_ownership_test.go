package repository_test

import (
	"testing"

	"judge-backend/internal/app/shared/model"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/image"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/infrastructure/repository"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/mocks"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// A real (in-memory sqlite) database on purpose: the defect under test is
// HOW the repository resolves which row to delete at call time — a mocked
// repository would only record that Delete was called with A's entity and
// could never show it landing on B's row.
func newSqliteImageRepository(t *testing.T) (*repository.ImageRepository, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=private"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Image{}))

	storagePathProvider := &mocks.StoragePathProvider{}
	storagePathProvider.On("GetStoragePath").Return("storage")
	r, ok := repository.NewImageRepository(db, storagePathProvider).(*repository.ImageRepository)
	require.True(t, ok)
	return r, db
}

func ownershipImage(t *testing.T, description string) *image.Image {
	t.Helper()
	img, err := image.New(&image.Config{
		Tag:         "ownership:latest",
		MountPath:   "/var/www/html",
		Description: description,
		StoragePath: "storage",
	})
	require.NoError(t, err)
	return img
}

// TestImageRepository_LateCompensationMustNotDeleteRecreatedImage drives the
// create → delete → recreate interleaving:
//
//	A wins CreateExclusive → (A pauses mid-flow) → the user deletes A →
//	B wins CreateExclusive for the same tag → A's late compensation fires.
//
// A's compensation must not remove B's row. Deleting "A's image" by
// re-resolving the tag at call time resolves to whatever row CURRENTLY owns
// the tag — which is B's — so compensation needs a stable ownership token
// (the row ID CreateExclusive returned), not a tag lookup.
func TestImageRepository_LateCompensationMustNotDeleteRecreatedImage(t *testing.T) {
	r, _ := newSqliteImageRepository(t)

	// A creates and owns the tag.
	imageA := ownershipImage(t, "A's image")
	tokenA, err := r.CreateExclusive(imageA)
	require.NoError(t, err)

	// A pauses; the user deletes A's image (the legitimate by-tag path).
	require.NoError(t, r.Delete(imageA))

	// B recreates the same tag and owns it now.
	imageB := ownershipImage(t, "B's image")
	tokenB, err := r.CreateExclusive(imageB)
	require.NoError(t, err)
	require.NotEqual(t, tokenA, tokenB, "the recreated image must be a NEW row — otherwise tokens cannot distinguish owners")

	// A's late compensation fires — DeleteIfOwned with A's stale token must
	// be a no-op for the new owner, and its verdict must say so: false is
	// what tells the compensation it no longer owns the tag-derived project
	// path either.
	owned, err := r.DeleteIfOwned(tokenA)
	require.NoError(t, err)
	require.False(t, owned, "a stale token must report NOT owned — the files fence depends on this verdict")

	current, err := r.FindByTag("ownership:latest")
	require.NoError(t, err)
	require.NotNil(t, current, "A's late compensation deleted B's freshly-created image")
	require.Equal(t, "B's image", current.Description(), "the surviving row must be B's")
}

// TestImageRepository_DeleteIfOwned_RemovesOwnRow confirms the token path
// still performs the real deletion when the row IS the caller's.
func TestImageRepository_DeleteIfOwned_RemovesOwnRow(t *testing.T) {
	r, _ := newSqliteImageRepository(t)

	imageA := ownershipImage(t, "A's image")
	tokenA, err := r.CreateExclusive(imageA)
	require.NoError(t, err)

	owned, err := r.DeleteIfOwned(tokenA)
	require.NoError(t, err)
	require.True(t, owned, "deleting one's own live row must report owned — it licenses the file cleanup")

	current, err := r.FindByTag("ownership:latest")
	require.NoError(t, err)
	require.Nil(t, current, "the owner's own compensation must actually delete its row")

	// A second call with the same token: the row is already gone, so the
	// verdict flips to false without error.
	ownedAgain, err := r.DeleteIfOwned(tokenA)
	require.NoError(t, err)
	require.False(t, ownedAgain, "a repeated delete must not claim ownership again")
}

// TestImageRepository_DeleteIfOwned_ZeroTokenIsNoOp pins the zero-value guard:
// a caller that never obtained a token cannot delete anything, and the
// verdict is not-owned.
func TestImageRepository_DeleteIfOwned_ZeroTokenIsNoOp(t *testing.T) {
	r, _ := newSqliteImageRepository(t)

	owned, err := r.DeleteIfOwned(0)
	require.NoError(t, err)
	require.False(t, owned)
}

// ── rename-on-delete: a soft-deleted row must release its tag ──────────

// TestImageRepository_DeleteRenamesTheSoftDeletedTag: the dead row keeps
// existing (build history hangs off its id) but under a renamed tag, so on
// MySQL — where the unique tag index has no partial WHERE — the original tag
// is free for re-creation.
func TestImageRepository_DeleteRenamesTheSoftDeletedTag(t *testing.T) {
	r, db := newSqliteImageRepository(t)

	imageA := ownershipImage(t, "A's image")
	tokenA, err := r.CreateExclusive(imageA)
	require.NoError(t, err)

	require.NoError(t, r.Delete(imageA))

	// The live view no longer resolves the tag…
	current, err := r.FindByTag("ownership:latest")
	require.NoError(t, err)
	require.Nil(t, current)

	// …and the dead row carries the renamed tag with deleted_at set.
	dead := &model.Image{}
	require.NoError(t, db.Unscoped().First(dead, "id = ?", tokenA).Error)
	require.Equal(t, model.DeletedImageTag("ownership:latest", tokenA), dead.Tag)
	require.True(t, dead.DeletedAt.Valid, "the rename must not resurrect the row")
}

// TestImageRepository_RepeatedDeleteRecreateCyclesKeepDistinctDeadTags: two
// soft-deleted incarnations of the same base tag must not collide with each
// other either — the row id in the suffix is what keeps N dead rows unique
// under a plain UNIQUE(tag) index.
func TestImageRepository_RepeatedDeleteRecreateCyclesKeepDistinctDeadTags(t *testing.T) {
	r, db := newSqliteImageRepository(t)

	first := ownershipImage(t, "first incarnation")
	idFirst, err := r.CreateExclusive(first)
	require.NoError(t, err)
	require.NoError(t, r.Delete(first))

	second := ownershipImage(t, "second incarnation")
	idSecond, err := r.CreateExclusive(second)
	require.NoError(t, err, "the renamed dead row must not block re-creation")
	require.NoError(t, r.Delete(second))

	deadFirst, deadSecond := &model.Image{}, &model.Image{}
	require.NoError(t, db.Unscoped().First(deadFirst, "id = ?", idFirst).Error)
	require.NoError(t, db.Unscoped().First(deadSecond, "id = ?", idSecond).Error)
	require.NotEqual(t, deadFirst.Tag, deadSecond.Tag, "two dead rows of the same base tag must have distinct renamed tags")
}
