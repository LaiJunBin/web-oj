package repository_test

import (
	"errors"
	"testing"

	"judge-backend/internal/app/identity_domain/user/domain/entity/user"
	domainrepo "judge-backend/internal/app/identity_domain/user/domain/repository"
	infrarepo "judge-backend/internal/app/identity_domain/user/infrastructure/repository"
	"judge-backend/internal/app/identity_domain/user/infrastructure/service/user/roleconverter"
	sharedlogger "judge-backend/internal/app/shared/infrastructure/logger"
	"judge-backend/internal/app/shared/model"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// A real (in-memory sqlite) database on purpose — the defect under test is
// what the unique username index does to a soft-deleted row, which mocks
// cannot show. Note the existing user_repository_test.go uses sqlmock and is
// deliberately left alone.
func newSqliteUserRepository(t *testing.T) (domainrepo.UserRepository, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=private"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.User{}))

	logger := sharedlogger.NewMockLogger()
	return infrarepo.NewUserRepository(db, roleconverter.NewRoleConverter(logger), logger), db
}

func liveUser(t *testing.T, username string) *user.User {
	t.Helper()
	return user.MustNew("Some Name", username, "hashed_password")
}

// TestUserRepository_DeleteRenamesTheSoftDeletedUsername: the dead row
// keeps existing but under a renamed username, so the unqualified unique
// index — plain UNIQUE on every database, no partial-index WHERE anywhere —
// no longer blocks recreating the name.
func TestUserRepository_DeleteRenamesTheSoftDeletedUsername(t *testing.T) {
	r, db := newSqliteUserRepository(t)

	alice := liveUser(t, "alice")
	require.NoError(t, r.Save(alice))
	id := alice.GetID()

	require.NoError(t, r.Delete(alice))

	// The live view no longer resolves the name…
	found, err := r.FindByUsername("alice")
	require.NoError(t, err)
	require.Nil(t, found)

	// …and the dead row carries the renamed username with deleted_at set.
	dead := &model.User{}
	require.NoError(t, db.Unscoped().First(dead, "id = ?", id).Error)
	require.Equal(t, model.DeletedUsername("alice", id), dead.Username)
	require.True(t, dead.DeletedAt.Valid, "the rename must not resurrect the row")
}

// TestUserRepository_RecreateAfterDeleteSucceeds is the regression
// proper: this exact Save used to hit the unique index of the soft-deleted
// row and surface as a bare 500.
func TestUserRepository_RecreateAfterDeleteSucceeds(t *testing.T) {
	r, db := newSqliteUserRepository(t)

	first := liveUser(t, "alice")
	require.NoError(t, r.Save(first))
	idFirst := first.GetID()
	require.NoError(t, r.Delete(first))

	second := liveUser(t, "alice")
	require.NoError(t, r.Save(second), "the renamed dead row must not block re-creation")
	require.NoError(t, r.Delete(second))

	// Two dead incarnations of the same base name carry distinct suffixes.
	deadFirst, deadSecond := &model.User{}, &model.User{}
	require.NoError(t, db.Unscoped().First(deadFirst, "id = ?", idFirst).Error)
	require.NoError(t, db.Unscoped().First(deadSecond, "id = ?", second.GetID()).Error)
	require.NotEqual(t, deadFirst.Username, deadSecond.Username)
}

// TestUserRepository_SaveDuplicateLiveUsername_MapsToSentinel: with dead rows
// renamed out of the way, the only duplicate left is a LIVE row racing the
// usecase's scoped pre-check — and it must map to the friendly sentinel, not
// bubble up as a raw driver error (the old bare-500 path).
func TestUserRepository_SaveDuplicateLiveUsername_MapsToSentinel(t *testing.T) {
	r, _ := newSqliteUserRepository(t)

	require.NoError(t, r.Save(liveUser(t, "alice")))

	err := r.Save(liveUser(t, "alice"))
	require.Error(t, err)
	require.ErrorIs(t, err, domainrepo.ErrDuplicateUsername)
}

// TestUserRepository_SaveAll covers the roster-import path: re-importing a
// roster whose accounts were deleted must succeed (the renamed dead rows do
// not occupy the names), while a LIVE duplicate is reported as the
// existing-user case with the username in the message — not as a row-format
// failure.
func TestUserRepository_SaveAll(t *testing.T) {
	t.Run("re-import after delete succeeds", func(t *testing.T) {
		r, _ := newSqliteUserRepository(t)

		original := liveUser(t, "bob")
		require.NoError(t, r.Save(original))
		require.NoError(t, r.Delete(original))

		require.NoError(t, r.SaveAll([]*user.User{liveUser(t, "bob")}),
			"a deleted roster must be re-importable under the same usernames")
	})

	t.Run("a live duplicate is the existing-user case", func(t *testing.T) {
		r, _ := newSqliteUserRepository(t)
		require.NoError(t, r.Save(liveUser(t, "carol")))

		err := r.SaveAll([]*user.User{liveUser(t, "carol")})
		require.Error(t, err)
		var existing *domainrepo.SaveAllExistingUserError
		require.True(t, errors.As(err, &existing),
			"a live duplicate must be reported as SaveAllExistingUserError, got %T: %v", err, err)
	})
}
