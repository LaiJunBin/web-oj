package repository_test

import (
	"testing"
	"time"

	infrarepo "judge-backend/internal/app/judge_domain/judge/infrastructure/repository"
	sharedlogger "judge-backend/internal/app/shared/infrastructure/logger"
	"judge-backend/internal/app/shared/model"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// FindUsernameByID feeds SUBMIT_USERNAME to the judge containers, which use
// it to create and address per-seat resources. A soft-deleted row's stored
// username was renamed to free the unique index (model.DeletedUsername) —
// the repository must hand back the ORIGINAL name, or an in-flight
// submission judged after its account's deletion would look for resources
// under "alice__deleted_17".
func TestFindUsernameByID_RestoresTheOriginalNameOfDeletedUsers(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=private"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.User{}))
	r := infrarepo.NewUserRepository(db, sharedlogger.NewMockLogger())

	now := time.Now()
	seed := func(u *model.User) {
		require.NoError(t, db.Unscoped().Create(u).Error)
	}
	seed(&model.User{ID: 1, Username: "alice"})
	seed(&model.User{ID: 2, Username: model.DeletedUsername("bob", 2),
		Model: gorm.Model{DeletedAt: gorm.DeletedAt{Time: now, Valid: true}}})
	// A LIVE user whose real name contains the marker must pass through
	// untouched — stripping is gated on deleted_at, not on the string shape.
	seed(&model.User{ID: 3, Username: "eve__deleted_9"})

	got, err := r.FindUsernameByID(1)
	require.NoError(t, err)
	require.Equal(t, "alice", got, "a live user resolves as-is")

	got, err = r.FindUsernameByID(2)
	require.NoError(t, err)
	require.Equal(t, "bob", got, "a deleted user's ORIGINAL name is restored for the judge pipeline")

	got, err = r.FindUsernameByID(3)
	require.NoError(t, err)
	require.Equal(t, "eve__deleted_9", got, "a live marker-shaped name is not stripped")
}
