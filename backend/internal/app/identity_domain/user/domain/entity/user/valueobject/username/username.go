package username

import (
	"errors"
	"strings"
)

var (
	ErrInvalidUsernameLength = errors.New("username must be at least 3 characters long")
)

// DeletedUsernameMarker duplicates model.DeletedUsernameMarker — the domain
// layer does not import shared/model, so the value is pinned equal by a test
// instead. A username containing it could collide with a soft-deleted row's
// renamed username, turning that row's delete into a unique-index failure
// (a 500, and a user that can never be deleted).
const DeletedUsernameMarker = "__deleted_"

// ContainsDeletedMarker reports whether a username uses the rename-on-delete
// namespace. Deliberately NOT part of Validate(): hydration (the repository's
// toDomainUser → user.New) must never couple to the marker rule, so a future
// Unscoped read of a renamed dead row can't blow up on it — the rule is
// enforced only at the creation entry points (create_user's validate,
// import_users' row loop). SetUsername currently has no callers; a future
// username-change API would be a third entry point to guard.
func ContainsDeletedMarker(value string) bool {
	return strings.Contains(value, DeletedUsernameMarker)
}

type Username string

func New(username string) (Username, error) {
	u := Username(username)
	if err := u.Validate(); err != nil {
		return "", err
	}

	return u, nil
}

func (u Username) String() string {
	return string(u)
}

func (u Username) Equals(other Username) bool {
	return u == other
}

func (u Username) Validate() error {
	if len(u) < 3 {
		return ErrInvalidUsernameLength
	}

	return nil
}
